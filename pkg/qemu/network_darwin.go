package qemu

import (
	"fmt"
)

func buildNetwork(id string, network NetworkConfig, platform *PlatformConfig) ([]string, error) {
	if platform == nil || platform.Network == nil {
		return nil, fmt.Errorf("platform network configuration required")
	}

	darwinNet := platform.Network
	if darwinNet.Bridged != nil && darwinNet.Shared != nil {
		return nil, fmt.Errorf("network configuration: Bridged and Shared are mutually exclusive")
	}

	var netdevArgs string
	var netdevArgsErr error

	if darwinNet.Bridged != nil {
		netdevArgs, netdevArgsErr = darwinNet.Bridged.buildNetdevArgs(id)
	} else if darwinNet.Shared != nil {
		netdevArgs, netdevArgsErr = darwinNet.Shared.buildNetdevArgs(id)
	} else {
		return nil, fmt.Errorf("network configuration required: either Bridged or Shared must be set")
	}
	if netdevArgsErr != nil {
		return nil, netdevArgsErr
	}

	args := []string{
		"-netdev", netdevArgs,
		"-device", fmt.Sprintf("%s,netdev=%s,mac=%s,id=%s", network.Driver, id, network.Mac, id),
	}

	return args, nil
}
