package utils

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

func ParseMb(sizeStr string) (float64, error) {
	// Parse the string into bytes
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %v", err)
	}

	// Convert bytes to megabytes (1 MB = 1024 * 1024 bytes)
	megabytes := float64(bytes) / (1024 * 1024)
	return megabytes, nil
}

func FormatMb(megabytes uint32) string {
	// Convert megabytes to bytes (1 MB = 1024 * 1024 bytes)
	bytes := uint64(megabytes * 1024 * 1024)

	// Format bytes into a human-readable string
	humanReadable := humanize.Bytes(bytes)
	return humanReadable
}
