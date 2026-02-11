//go:build darwin

package qemu

import "fmt"

// VmnetBridged holds configuration for vmnet-bridged networking mode.
type VmnetBridged struct {
	Interface string // interface name for bridged networking
}

// buildNetdevArgs returns the QEMU netdev arguments for bridged mode.
func (b *VmnetBridged) buildNetdevArgs(id string) (string, error) {
	if b.Interface == "" {
		return "", fmt.Errorf("vmnet-bridged: Interface must be set")
	}
	return fmt.Sprintf("vmnet-bridged,id=%s,ifname=%s", id, b.Interface), nil
}

// VmnetShared holds configuration for vmnet-shared networking mode.
type VmnetShared struct {
	StartAddress string // e.g., "192.168.33.1"
	EndAddress   string // e.g., "192.168.33.254"
	SubnetMask   string // e.g., "255.255.255.0"
}

// buildNetdevArgs returns the QEMU netdev arguments for shared mode.
// If all address fields are empty, a bare vmnet-shared netdev is returned (QEMU defaults).
// If all address fields are set, they are included in the netdev args.
// If only some fields are set, an error is returned.
func (s *VmnetShared) buildNetdevArgs(id string) (string, error) {
	fields := []string{s.StartAddress, s.EndAddress, s.SubnetMask}
	setCount := 0
	for _, f := range fields {
		if f != "" {
			setCount++
		}
	}

	switch setCount {
	case 0:
		return fmt.Sprintf("vmnet-shared,id=%s", id), nil
	case len(fields):
		return fmt.Sprintf("vmnet-shared,id=%s,start-address=%s,end-address=%s,subnet-mask=%s",
			id, s.StartAddress, s.EndAddress, s.SubnetMask), nil
	default:
		return "", fmt.Errorf("vmnet-shared: all of StartAddress, EndAddress, and SubnetMask must be set together or all left empty")
	}
}

// DarwinNetworkConfig holds macOS-specific network configuration.
// Either Bridged or Shared must be set (but not both).
type DarwinNetworkConfig struct {
	Bridged *VmnetBridged // for vmnet-bridged mode
	Shared  *VmnetShared  // for vmnet-shared mode
}
