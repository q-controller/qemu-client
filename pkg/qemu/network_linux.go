package qemu

import "fmt"

func build_network(id string, network NetworkConfig) ([]string, error) {
	args := []string{}

	args = append(args, "-device", fmt.Sprintf("%s,netdev=net0,mac=%s,id=%s", network.Driver, network.Mac, id))
	args = append(args, "-netdev", fmt.Sprintf("tap,id=net0,ifname=%s,script=no,downscript=no", id))

	return args, nil
}
