//go:build darwin

package qemu

// PlatformConfig holds macOS-specific configuration.
type PlatformConfig struct {
	Network *DarwinNetworkConfig
}
