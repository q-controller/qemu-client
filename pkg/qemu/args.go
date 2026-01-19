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
	Qmp         string
	Qga         string
	Image       string
	CloudInit   CloudInitConfig
	Hardware    Hardware
	TmpDir      string
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

func Qmp(path string) Option {
	return func(config *QemuConfig) {
		config.Qmp = path
	}
}

func Qga(path string) Option {
	return func(config *QemuConfig) {
		config.Qga = path
	}
}

func Image(path string) Option {
	return func(config *QemuConfig) {
		config.Image = path
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

func TmpDir(path string) Option {
	return func(config *QemuConfig) {
		config.TmpDir = path
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

	image := utils.Image{
		Path: config.Image,
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

	netArgs, netArgsErr := build_network(config.Id, config.Network)
	if netArgsErr != nil {
		return nil, netArgsErr
	}
	args = append(args, netArgs...)

	args = append(args, "-qmp", fmt.Sprintf("unix:%s,server,wait=off", config.Qmp))
	args = append(args, "-cpu", "host")
	args = append(args, "-smp", fmt.Sprintf("%d", config.Hardware.Cpus))
	args = append(args, "-hda", config.Image)
	args = append(args, "-device", "virtio-serial")
	args = append(args, "-chardev", fmt.Sprintf("socket,path=%s,server=on,wait=off,id=charchannel0", config.Qga))
	args = append(args, "-device", "virtserialport,chardev=charchannel0,name=org.qemu.guest_agent.0")

	tmpDir, tmpDirErr := os.MkdirTemp("", "cloudinit-*")
	if tmpDirErr != nil {
		return nil, tmpDirErr
	}

	cloudInitPath, cloudInitErr := utils.CreateCloudInitISO(config.CloudInit.Userdata, config.CloudInit.NetworkConfig, tmpDir, config.Id)
	if cloudInitErr != nil {
		return nil, cloudInitErr
	}
	args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw,if=virtio", cloudInitPath))

	if config.Bios != "" {
		args = append(args, "-bios", config.Bios)
	}

	return args, nil
}
