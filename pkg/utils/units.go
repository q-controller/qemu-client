package utils

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
)

func ParseMb(sizeStr string) (uint64, error) {
	// Parse the string into bytes
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %v", err)
	}

	return BytesToMb(bytes), nil
}

// Convert bytes to megabytes
func BytesToMb(bytes uint64) uint64 {
	megabytes := float64(bytes) / humanize.MiByte
	return uint64(math.Ceil(megabytes))
}

func MbToBytes(mb uint64) uint64 {
	return mb * humanize.MiByte
}

func FormatMb(megabytes uint32) string {
	return fmt.Sprintf("%dM", megabytes)
}

func FormatBytes(bytes uint64) string {
	humanReadable := humanize.SI(float64(bytes), "B")
	return humanReadable
}
