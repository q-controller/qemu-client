package utils

import (
	"fmt"
	"runtime"
)

func GetBios() (string, error) {
	switch runtime.GOARCH {
	case "arm64":
		return "edk2-aarch64-code.fd", nil
	case "amd64":
		return "", nil
	}
	return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
}
