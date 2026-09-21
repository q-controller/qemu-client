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

// Notification carries a chunk of new log output for a watched stream. Reset is
// true for the first chunk read from the start of a file that was truncated or
// replaced since the previous chunk.
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

// cursor is the read position within a watched path together with the
// identity of the file it was taken from.
type cursor struct {
	offset int64
	file   os.FileInfo
}

// drainOffset reads bytes from the file at path starting at cur.offset into the
// caller-provided buffer p, reading until p is full, EOF is reached, or ctx is
// cancelled. It returns the number of bytes read and the cursor to continue
// from next time; io.EOF means the end of file was reached (with n < len(p)).
// On cancellation it returns (0, cur, ctx.Err()); the call consumes nothing,
// so the cursor is left untouched.
// If the file at path is not the one the cursor was taken from (rotation) or
// has shrunk below the offset (in-place truncation), reading restarts from the
// beginning, so the returned offset may be lower than the one passed in.
func drainOffset(ctx context.Context, path string, p []byte, cur cursor) (int, cursor, error) {
	orig := cur
	f, err := os.Open(path)
	if err != nil {
		return 0, orig, err
	}
	defer f.Close()

	if info, statErr := f.Stat(); statErr == nil {
		if (cur.file != nil && !os.SameFile(cur.file, info)) || info.Size() < cur.offset {
			cur.offset = 0
		}
		cur.file = info
	}

	if _, err := f.Seek(cur.offset, io.SeekStart); err != nil {
		return 0, orig, err
	}

	total := 0
	for total < len(p) {
		// Bail out between reads if the caller cancelled; nothing consumed yet.
		select {
		case <-ctx.Done():
			return 0, orig, ctx.Err()
		default:
		}

		n, err := f.Read(p[total:])
		total += n
		if err != nil {
			cur.offset += int64(total)
			return total, cur, err // includes io.EOF
		}
	}
	cur.offset += int64(total)
	return total, cur, nil
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

	events := make(chan Kind)
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
				select {
				case events <- kind:
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
		cursors := make(map[string]cursor)

		// drain reads path to EOF, emitting a notification per chunk. It
		// returns false when ctx is cancelled so the caller stops.
		drain := func(kind Kind, path string) bool {
			for {
				buffer := make([]byte, 4096)
				cur := cursors[path]
				n, next, err := drainOffset(ctx, path, buffer, cur)
				cursors[path] = next
				if n > 0 {
					// A chunk that starts at 0 while the cursor was further on
					// means drainOffset rewound.
					reset := cur.offset > 0 && next.offset == int64(n)
					select {
					case notifications <- Notification{Kind: kind, Data: buffer[:n], Reset: reset}:
					case <-ctx.Done():
						return false
					}
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
		if !drain(KindStdout, stdout) || !drain(KindStderr, stderr) {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case kind, ok := <-events:
				if !ok {
					return
				}
				var path string
				switch kind {
				case KindStdout:
					path = stdout
				case KindStderr:
					path = stderr
				default:
					continue
				}
				if !drain(kind, path) {
					return
				}
			}
		}
	}()

	return &LogTailer{watcher: watcher, notifications: notifications}, nil
}
