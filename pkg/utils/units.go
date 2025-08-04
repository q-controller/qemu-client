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

// Convert bytes to megabytes (1 MB = 1024 * 1024 bytes)
func BytesToMb(bytes uint64) uint64 {
	megabytes := float64(bytes) / (1024 * 1024)
	return uint64(math.Ceil(megabytes/10) * 10)
}

func MbToBytes(mb uint64) uint64 {
	return mb * (1024 * 1024)
}

func FormatMb(megabytes uint32) string {
	// Convert megabytes to bytes (1 MB = 1024 * 1024 bytes)
	bytes := uint64(megabytes * 1024 * 1024)

	// Format bytes into a human-readable string
	humanReadable := humanize.Bytes(bytes)
	return humanReadable
}
