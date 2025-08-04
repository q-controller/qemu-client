package qemu

import (
	"fmt"
	"os/exec"
	"strings"
)

func getDefaultInterface() (string, error) {
	out, err := exec.Command("route", "get", "default").Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "interface:") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				return strings.TrimSpace(fields[1]), nil
			}
		}
	}
	return "", fmt.Errorf("default interface not found")
}

func build_network(id string, network NetworkConfig) ([]string, error) {
	ifc, ifcErr := getDefaultInterface()
	if ifcErr != nil {
		return nil, ifcErr
	}

	args := []string{}

	args = append(args, "-device", fmt.Sprintf("%s,netdev=%s,mac=%s", network.Driver, id, network.Mac))
	args = append(args, "-netdev", fmt.Sprintf("vmnet-bridged,id=%s,ifname=%s", id, ifc))

	return args, nil
}
