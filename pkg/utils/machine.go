package utils

import (
	"fmt"
	"runtime"
)

func GetMachineType() (string, error) {
	switch runtime.GOARCH {
	case "arm64", "arm":
		return "virt", nil
	case "amd64":
		return "q35", nil
	}
	return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
}
