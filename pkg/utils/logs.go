package utils

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindStdout
	KindStderr
)

func (k Kind) String() string {
	switch k {
	case KindStdout:
		return "stdout"
	case KindStderr:
		return "stderr"
	default:
		return "unknown"
	}
}

// Data is an internal change signal from the file watcher to the reader. Reset
// is true when the stream was rotated (renamed or removed) and reading should
// restart from the start of the file.
type Data struct {
	Kind  Kind
	Reset bool
}

// Notification carries a chunk of new log output for a watched stream. Reset is
// true for the first chunk after a rotation, signalling the consumer to discard
// what it has shown so far for that stream.
type Notification struct {
	Kind  Kind
	Data  []byte
	Reset bool
}

type LogTailer struct {
	watcher       *fsnotify.Watcher
	notifications chan Notification
}

func (w *LogTailer) Close() error {
	return w.watcher.Close()
}

// drainOffset reads bytes from the file at path starting at offset into the
// caller-provided buffer p, reading until p is full, EOF is reached, or ctx is
// cancelled. It returns the number of bytes read and the offset to continue
// from next time; io.EOF means the end of file was reached (with n < len(p)).
// On cancellation it returns (0, offset, ctx.Err()); the call consumes nothing,
// so the offset is left untouched.
//
// If the file has shrunk below offset (in-place truncation), reading restarts
// from the beginning, so the returned offset may be lower than the one passed
// in. The caller should store the returned offset rather than computing
// offset+n itself.
func drainOffset(ctx context.Context, path string, p []byte, offset int64) (int, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, offset, err
	}
	defer f.Close()

	// If the file shrank below offset it was truncated; restart from 0.
	if info, statErr := f.Stat(); statErr == nil && info.Size() < offset {
		offset = 0
	}

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return 0, offset, err
	}

	total := 0
	for total < len(p) {
		// Bail out between reads if the caller cancelled; nothing consumed yet.
		select {
		case <-ctx.Done():
			return 0, offset, ctx.Err()
		default:
		}

		n, err := f.Read(p[total:])
		total += n
		if err != nil {
			return total, offset + int64(total), err // includes io.EOF
		}
	}
	return total, offset + int64(total), nil
}

// Notifications returns the channel of log chunks for the watched streams. It
// closes when the context passed to NewLogTailer is cancelled or the LogTailer
// is closed.
func (w *LogTailer) Notifications() <-chan Notification {
	return w.notifications
}

func NewLogTailer(ctx context.Context, stdout, stderr string) (*LogTailer, error) {
	watcher, watcherErr := fsnotify.NewWatcher()
	if watcherErr != nil {
		return nil, watcherErr
	}

	// Watch the parent directories, not the files: a per-file watch is dropped
	// when a rotation replaces the file, whereas a directory watch keeps seeing
	// writes and the new file's creation, so following survives rotation.
	// (stdout and stderr usually share a directory; re-adding it is a no-op.)
	for _, p := range []string{stdout, stderr} {
		if addErr := watcher.Add(filepath.Dir(p)); addErr != nil {
			_ = watcher.Close()
			return nil, addErr
		}
	}

	events := make(chan Data)
	notifications := make(chan Notification)
	go func() {
		// Closing the watcher on exit makes ctx-cancel a complete teardown;
		// the explicit Close stays an optional early release (idempotent).
		defer func() { _ = watcher.Close() }()
		defer close(events)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				var kind Kind
				switch event.Name {
				case stdout:
					kind = KindStdout
				case stderr:
					kind = KindStderr
				default:
					continue
				}
				// A plain write just appends; anything else (rename,
				// remove, chmod, ...) means the stream should be re-read
				// from the start.
				data := Data{Kind: kind, Reset: !event.Has(fsnotify.Write)}
				select {
				case events <- data:
				case <-ctx.Done():
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("log watcher error", "error", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		defer close(notifications)
		offsets := make(map[string]int64)

		// drain reads path to EOF, emitting a notification per chunk. It
		// returns false when ctx is cancelled so the caller stops.
		drain := func(kind Kind, path string, reset bool) bool {
			if reset {
				offsets[path] = 0
			}
			for {
				buffer := make([]byte, 4096)
				offset := offsets[path]
				n, newOffset, err := drainOffset(ctx, path, buffer, offset)
				offsets[path] = newOffset
				if n > 0 {
					// newOffset < offset means drainOffset rewound (in-place
					// truncation); tell the consumer to clear before appending.
					select {
					case notifications <- Notification{Kind: kind, Data: buffer[:n], Reset: reset || newOffset < offset}:
					case <-ctx.Done():
						return false
					}
					reset = false // only the first chunk of a reset clears the view
					continue
				}
				if ctx.Err() != nil {
					return false
				}
				if err != nil && !errors.Is(err, io.EOF) && !os.IsNotExist(err) {
					slog.Error("failed to read log file", "path", path, "error", err)
				}
				return true
			}
		}

		// Emit the existing contents before following new output.
		if !drain(KindStdout, stdout, false) || !drain(KindStderr, stderr, false) {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				var path string
				switch event.Kind {
				case KindStdout:
					path = stdout
				case KindStderr:
					path = stderr
				default:
					continue
				}
				if !drain(event.Kind, path, event.Reset) {
					return
				}
			}
		}
	}()

	return &LogTailer{watcher: watcher, notifications: notifications}, nil
}
