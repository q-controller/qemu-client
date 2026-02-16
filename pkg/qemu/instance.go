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
	Cpus      uint32
	Memory    uint32 // in MB
	Disk      uint32 // in MB
	HwAddr    string
	CloudInit CloudInitConfig
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
		os.Remove(qmpSocketFor(name))
		os.Remove(qgaSocketFor(name))
	}()

	return &Instance{
		QMP:  qmpSocketFor(name),
		QGA:  qgaSocketFor(name),
		Pid:  pid,
		Done: ch,
	}, nil
}

func Start(name, url, outFilePath, errFilePath string, config Config) (*Instance, error) {
	qemuBinary, qemuBinaryErr := utils.GetQemuBinary()
	if qemuBinaryErr != nil {
		return nil, qemuBinaryErr
	}

	if _, err := exec.LookPath(qemuBinary); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", qemuBinary, qemuBinary)
	}

	tmpDir, tmpDirErr := os.MkdirTemp("", "cloudinit-*")
	if tmpDirErr != nil {
		return nil, tmpDirErr
	}

	machineType, machineTypeErr := utils.GetMachineType()
	if machineTypeErr != nil {
		return nil, machineTypeErr
	}

	bios, biosErr := utils.GetBios()
	if biosErr != nil {
		return nil, biosErr
	}

	args, argsErr := BuildQemuArgs(
		Id(name),
		Machine(machineType),
		Accelerator(utils.GetAccelerator()),
		Memory(config.Memory),
		Disk(config.Disk),
		Cpus(int(config.Cpus)),
		Network(NetworkConfig{
			Mac:    config.HwAddr,
			Driver: "virtio-net",
		}),
		Image(url),
		CloudInit(config.CloudInit),
		TmpDir(tmpDir),
		Bios(bios),
		Qmp(qmpSocketFor(name)),
		Qga(qgaSocketFor(name)),
	)
	if argsErr != nil {
		return nil, argsErr
	}

	// Remove stale socket files from a previous run before starting QEMU.
	// QEMU creates these as server sockets; if stale files remain, the new
	// instance may fail to bind or clients may get "connection refused".
	os.Remove(qmpSocketFor(name))
	os.Remove(qgaSocketFor(name))

	slog.Info("QEMU command", "binary", qemuBinary, "args", args)
	command := exec.Command(qemuBinary, args...)
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
		os.Remove(qmpSocketFor(name))
		os.Remove(qgaSocketFor(name))
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
