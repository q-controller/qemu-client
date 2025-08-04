package qemu

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/q-controller/qemu-client/pkg/utils"
)

type Instance struct {
	QMP  string
	QGA  string
	Pid  int
	Done <-chan interface{}
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

func Attach(name string, pid int) (*Instance, error) {
	proc, procErr := os.FindProcess(pid)
	if procErr != nil {
		return nil, procErr
	}

	ch := make(chan interface{})

	go func() {
		defer close(ch)

		for {
			err := proc.Signal(syscall.Signal(0)) // no-op signal
			if err != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		ch <- true
	}()

	return &Instance{
		QMP:  qmpSocketFor(name),
		QGA:  qgaSocketFor(name),
		Pid:  pid,
		Done: ch,
	}, nil
}

func Start(name, url, outFilePath, errFilePath string, config Config) (*Instance, error) {
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
		Disk(config.Disk),
		Cpus(int(config.Cpus)),
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
	outFile, outFileErr := os.OpenFile(outFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if outFileErr != nil {
		return nil, outFileErr
	}
	command.Stdout = outFile

	errFile, errFileErr := os.OpenFile(errFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if errFileErr != nil {
		return nil, errFileErr
	}
	command.Stderr = errFile

	// Detach from parent process
	command.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// Start QEMU non-blocking
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("failed to execute QEMU: %w", err)
	}
	slog.Debug("QEMU VM started", "pid", command.Process.Pid)

	ch := make(chan interface{})

	go func(name string) {
		defer os.RemoveAll(tmpDir)
		defer close(ch)

		waitErr := command.Wait()
		if waitErr != nil {
			slog.Info("Exited with error", "error", waitErr)
		}
		ch <- true
	}(name)

	return &Instance{
		QMP:  qmpSocketFor(name),
		QGA:  qgaSocketFor(name),
		Pid:  command.Process.Pid,
		Done: ch,
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
