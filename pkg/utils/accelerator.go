package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

// GetAccelerator returns the qemu -accel value for the current host.
// On Linux it requires /dev/kvm; on Darwin it requires the Hypervisor
// framework (assumed available on supported macOS versions; not probed).
//
// allowEmulationFallback is honoured only on Linux. When KVM is missing
// and the flag is true, "tcg" is returned with a warning log; when false,
// an error is returned so production hosts fail loudly instead of
// silently degrading. On Darwin the flag is currently ignored because
// HVF has no cheap availability probe; if HVF is genuinely unavailable,
// qemu surfaces the error at start.
func GetAccelerator(allowEmulationFallback bool) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "hvf", nil
	case "linux":
		if _, err := os.Stat("/dev/kvm"); err == nil {
			return "kvm", nil
		}
		if allowEmulationFallback {
			slog.Warn("KVM unavailable; falling back to TCG software emulation")
			return "tcg", nil
		}
		return "", errors.New("/dev/kvm not available and allow_emulation_fallback is disabled")
	}
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}
