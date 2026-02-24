package qemu

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	Platform  *PlatformConfig // platform-specific configuration
	CloudInit CloudInitConfig
}

// Path helpers — all runtime files live inside the instance directory.

func QmpSocketPath(dir string) string {
	return filepath.Join(dir, "qmp.sock")
}

func QgaSocketPath(dir string) string {
	return filepath.Join(dir, "qga.sock")
}

func PidfilePath(dir string) string {
	return filepath.Join(dir, "pid")
}

func StdoutPath(dir string) string {
	return filepath.Join(dir, "stdout")
}

func StderrPath(dir string) string {
	return filepath.Join(dir, "stderr")
}

func ImagePath(dir string) string {
	return filepath.Join(dir, "image")
}

func CloudInitPath(dir string) string {
	return filepath.Join(dir, "cloudinit")
}

func ReadPidfile(dir string) (int, error) {
	data, err := os.ReadFile(PidfilePath(dir))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pidfile content: %w", err)
	}
	return pid, nil
}

func ProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func Attach(name, dir string, pid int) (*Instance, error) {
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
	}()

	return &Instance{
		QMP:  QmpSocketPath(dir),
		QGA:  QgaSocketPath(dir),
		Pid:  pid,
		Done: ch,
	}, nil
}

func Start(name, dir string, config Config) (*Instance, error) {
	qemuBinary, qemuBinaryErr := utils.GetQemuBinary()
	if qemuBinaryErr != nil {
		return nil, qemuBinaryErr
	}

	if _, err := exec.LookPath(qemuBinary); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", qemuBinary, qemuBinary)
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
		Platform(config.Platform),
		Dir(dir),
		CloudInit(config.CloudInit),
		Bios(bios),
	)
	if argsErr != nil {
		return nil, argsErr
	}

	// Remove stale socket files from a previous run before starting QEMU.
	os.Remove(QmpSocketPath(dir))
	os.Remove(QgaSocketPath(dir))

	slog.Info("QEMU command", "binary", qemuBinary, "args", args)
	command := exec.Command(qemuBinary, args...)

	outFile, outFileErr := os.OpenFile(StdoutPath(dir), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if outFileErr != nil {
		return nil, outFileErr
	}
	command.Stdout = outFile

	errFile, errFileErr := os.OpenFile(StderrPath(dir), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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

	go func() {
		defer close(ch)

		waitErr := command.Wait()
		if waitErr != nil {
			slog.Info("Exited with error", "error", waitErr)
		}
		ch <- true
	}()

	return &Instance{
		QMP:  QmpSocketPath(dir),
		QGA:  QgaSocketPath(dir),
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
