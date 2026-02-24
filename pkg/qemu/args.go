package qemu

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/q-controller/qemu-client/pkg/utils"
)

type NetworkConfig struct {
	Driver string
	Mac    string
}

type Hardware struct {
	Memory uint32 // in MB
	Disk   uint32 // in MB
	Cpus   int
}

type CloudInitConfig struct {
	Userdata      string
	NetworkConfig string
}

type QemuConfig struct {
	Id          string
	Machine     string
	Accelerator string
	Network     NetworkConfig
	Platform    *PlatformConfig
	Dir         string // instance directory — all runtime paths derived from this
	CloudInit   CloudInitConfig
	Hardware    Hardware
	Bios        string
}

type Option func(*QemuConfig)

func Id(id string) Option {
	return func(config *QemuConfig) {
		config.Id = id
	}
}

func Machine(machine string) Option {
	return func(config *QemuConfig) {
		config.Machine = machine
	}
}

func Accelerator(accel string) Option {
	return func(config *QemuConfig) {
		config.Accelerator = accel
	}
}

func Network(network NetworkConfig) Option {
	return func(config *QemuConfig) {
		config.Network = network
	}
}

func Platform(platform *PlatformConfig) Option {
	return func(config *QemuConfig) {
		config.Platform = platform
	}
}

func Dir(path string) Option {
	return func(config *QemuConfig) {
		config.Dir = path
	}
}

func CloudInit(cloudinit CloudInitConfig) Option {
	return func(config *QemuConfig) {
		config.CloudInit = cloudinit
	}
}

func Memory(memory uint32) Option {
	return func(config *QemuConfig) {
		config.Hardware.Memory = memory
	}
}

func Disk(disk uint32) Option {
	return func(config *QemuConfig) {
		config.Hardware.Disk = disk
	}
}

func Cpus(cpus int) Option {
	return func(config *QemuConfig) {
		config.Hardware.Cpus = cpus
	}
}

func Bios(bios string) Option {
	return func(config *QemuConfig) {
		config.Bios = bios
	}
}

func BuildQemuArgs(opts ...Option) ([]string, error) {
	config := &QemuConfig{
		Machine: "q35",
		Network: NetworkConfig{
			Driver: "virtio-net",
		},
		Hardware: Hardware{
			Memory: 1024,      // 1 GB
			Disk:   40 * 1024, // 40 GB
			Cpus:   1,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	imagePath := ImagePath(config.Dir)
	qmpPath := QmpSocketPath(config.Dir)
	qgaPath := QgaSocketPath(config.Dir)
	pidfilePath := PidfilePath(config.Dir)

	image := utils.Image{
		Path: imagePath,
	}

	if info, infoErr := image.Info(); infoErr == nil {
		if utils.BytesToMb(info.VirtualSizeBytes) < uint64(config.Hardware.Disk) {
			if resizeErr := image.Resize(utils.MbToBytes(uint64(config.Hardware.Disk))); resizeErr != nil {
				return nil, resizeErr
			}
		}
	} else {
		slog.Error("Failed to get image info", "error", infoErr)
	}

	args := []string{}

	args = append(args, "-machine", config.Machine)
	args = append(args, "-accel", config.Accelerator)
	args = append(args, "-m", utils.FormatMb(config.Hardware.Memory))
	args = append(args, "-nographic")

	netArgs, netArgsErr := buildNetwork(config.Id, config.Network, config.Platform)
	if netArgsErr != nil {
		return nil, netArgsErr
	}
	args = append(args, netArgs...)

	args = append(args, "-qmp", fmt.Sprintf("unix:%s,server,wait=off", qmpPath))
	args = append(args, "-cpu", "host")
	args = append(args, "-smp", fmt.Sprintf("%d", config.Hardware.Cpus))
	args = append(args, "-hda", imagePath)
	args = append(args, "-pidfile", pidfilePath)
	args = append(args, "-device", "virtio-serial")
	args = append(args, "-chardev", fmt.Sprintf("socket,path=%s,server=on,wait=off,id=charchannel0", qgaPath))
	args = append(args, "-device", "virtserialport,chardev=charchannel0,name=org.qemu.guest_agent.0")

	cloudInitDir := CloudInitPath(config.Dir)
	if mkdirErr := os.MkdirAll(cloudInitDir, 0755); mkdirErr != nil {
		return nil, mkdirErr
	}

	cloudInitPath, cloudInitErr := utils.CreateCloudInitISO(config.CloudInit.Userdata, config.CloudInit.NetworkConfig, cloudInitDir, config.Id)
	if cloudInitErr != nil {
		return nil, cloudInitErr
	}
	args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw,if=virtio", cloudInitPath))

	if config.Bios != "" {
		args = append(args, "-bios", config.Bios)
	}

	args = append(args, "-device", fmt.Sprintf("virtio-balloon,id=balloon-%s,guest-stats-polling-interval=2", config.Id))

	return args, nil
}
