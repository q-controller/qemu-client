package qemu

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/q-controller/qemu-client/pkg/utils"
)

type Instance struct {
	QMP    string
	QGA    string
	Pid    int
	Done   <-chan interface{}
	Stderr <-chan string
	Stdout <-chan string
}

type Config struct {
	Cpus     uint32
	Memory   string
	Disk     string
	UserData string
	HwAddr   string
}

func qmpSocketFor(name string) string {
	return fmt.Sprintf("/tmp/%s.sock", name)
}

func qgaSocketFor(name string) string {
	return fmt.Sprintf("/tmp/qga-%s.sock", name)
}

func Start(name, url string, config Config) (*Instance, error) {
	const QEMU = "qemu-system-x86_64"
	if _, err := exec.LookPath(QEMU); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", QEMU, QEMU)
	}

	tmpDir, tmpDirErr := os.MkdirTemp("", "cloudinit-*")
	if tmpDirErr != nil {
		return nil, tmpDirErr
	}

	args, argsErr := BuildQemuArgs(
		Id(name),
		Machine("q35"),
		Accelerator(utils.GetAccelerator()),
		Memory(config.Memory),
		Network(NetworkConfig{
			Mac:    config.HwAddr,
			Driver: "virtio-net",
		}),
		Image(url),
		Userdata(config.UserData),
		TmpDir(tmpDir),
	)
	if argsErr != nil {
		return nil, argsErr
	}

	command := exec.Command(QEMU, args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	errCh := make(chan string)
	outCh := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if text := strings.TrimSpace(scanner.Text()); text != "" {
				outCh <- text
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Error("Error reading QEMU stdout", "error", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if text := strings.TrimSpace(scanner.Text()); text != "" {
				errCh <- text
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Error("Error reading QEMU stderr", "error", err)
		}
	}()

	// Start QEMU non-blocking
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("failed to execute QEMU: %w", err)
	}
	slog.Debug("QEMU VM started", "pid", command.Process.Pid)

	ch := make(chan interface{})

	go func(name string) {
		defer os.RemoveAll(tmpDir)
		defer close(ch)
		defer close(outCh)
		defer close(errCh)

		waitErr := command.Wait()
		if waitErr != nil {
			slog.Info("Exited with error", "error", waitErr)
		}
		ch <- true
	}(name)

	return &Instance{
		QMP:    qmpSocketFor(name),
		QGA:    qgaSocketFor(name),
		Pid:    command.Process.Pid,
		Done:   ch,
		Stderr: errCh,
		Stdout: outCh,
	}, nil
}

func (i *Instance) Stop() error {
	proc, err := os.FindProcess(i.Pid)
	if err != nil {
		return err
	}

	// Send SIGTERM for graceful termination
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	return nil
}
