package utils

import (
	"runtime"
)

func GetAccelerator() string {
	switch runtime.GOOS {
	case "darwin":
		return "hvf"
	case "linux":
		return "kvm"
	}
	return ""
}
