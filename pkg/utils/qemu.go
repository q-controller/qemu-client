package utils

import (
	"fmt"
	"runtime"
)

func GetQemuBinary() (string, error) {
	switch runtime.GOARCH {
	case "arm64":
		return "qemu-system-aarch64", nil
	case "amd64":
		return "qemu-system-x86_64", nil
	}

	return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
}
