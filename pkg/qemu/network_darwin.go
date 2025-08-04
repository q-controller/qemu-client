package qemu

import "fmt"

func build_network(id string, network NetworkConfig) ([]string, error) {
	args := []string{}

	args = append(args, "-device", fmt.Sprintf("%s,netdev=%s,mac=%s", network.Driver, id, network.Mac))
	args = append(args, "-netdev", fmt.Sprintf("vmnet-shared,id=%s", id))

	return args, nil
}
