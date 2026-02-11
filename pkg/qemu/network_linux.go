package qemu

import "fmt"

func buildNetwork(id string, network NetworkConfig, platform *PlatformConfig) ([]string, error) {
	args := []string{}

	args = append(args, "-device", fmt.Sprintf("%s,netdev=%s,mac=%s,id=%s", network.Driver, id, network.Mac, id))
	args = append(args, "-netdev", fmt.Sprintf("tap,id=%s,ifname=%s,script=no,downscript=no", id, id))

	return args, nil
}
