//go:build linux

package qemu

// PlatformConfig holds Linux-specific configuration.
type PlatformConfig struct {
	Network *LinuxNetworkConfig
}
