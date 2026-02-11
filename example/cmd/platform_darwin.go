//go:build darwin

package cmd

import "github.com/q-controller/qemu-client/pkg/qemu"

// getPlatformConfig returns macOS-specific platform configuration.
// This example uses vmnet-shared mode with a private network.
func getPlatformConfig() (*qemu.PlatformConfig, error) {
	return &qemu.PlatformConfig{
		Network: &qemu.DarwinNetworkConfig{
			Shared: &qemu.VmnetShared{
				StartAddress: "192.168.100.1",
				EndAddress:   "192.168.100.254",
				SubnetMask:   "255.255.255.0",
			},
		},
	}, nil
}
