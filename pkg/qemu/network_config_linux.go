//go:build linux

package qemu

// LinuxNetworkConfig holds Linux-specific network configuration.
// Currently empty as Linux tap networking uses the VM ID as interface name.
type LinuxNetworkConfig struct {
	// Future Linux-specific network fields can be added here
}
