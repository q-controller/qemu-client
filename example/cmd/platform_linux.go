//go:build linux

package cmd

import "github.com/q-controller/qemu-client/pkg/qemu"

// getPlatformConfig returns Linux-specific platform configuration.
// Linux uses tap networking, which doesn't require additional configuration.
func getPlatformConfig() (*qemu.PlatformConfig, error) {
	return &qemu.PlatformConfig{
		Network: &qemu.LinuxNetworkConfig{},
	}, nil
}
