package qemu

import "fmt"

func buildNetwork(id string, network NetworkConfig, platform *PlatformConfig) ([]string, error) {
	return []string{
		"-device", fmt.Sprintf("%s,netdev=%s,mac=%s,id=%s", network.Driver, id, network.Mac, id),
		"-netdev", fmt.Sprintf("tap,id=%s,ifname=%s,script=no,downscript=no", id, id),
	}, nil
}
