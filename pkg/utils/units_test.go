package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMb(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
		wantErr  bool
	}{
		{
			name:     "1MB",
			input:    "1MB",
			expected: 1, // Rounded up to nearest MB
			wantErr:  false,
		},
		{
			name:     "500KB",
			input:    "500KB",
			expected: 1, // 0.5MB rounded up to nearest MB
			wantErr:  false,
		},
		{
			name:     "1GB",
			input:    "1GB",
			expected: 954, // 1GB (1,000,000,000 bytes) / (1024*1024) = 953.67... → 954
			wantErr:  false,
		},
		{
			name:     "2.5GB",
			input:    "2.5GB",
			expected: 2385, // 2.5GB (2,500,000,000 bytes) / (1024*1024) = 2384.18... → 2385
			wantErr:  false,
		},
		{
			name:     "15MB",
			input:    "15MB",
			expected: 15, // Rounded up to nearest MB
			wantErr:  false,
		},
		{
			name:     "1048576B",
			input:    "1048576B",
			expected: 1, // 1MiB in binary
			wantErr:  false,
		},
		{
			name:     "invalid input",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMb(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "ParseMb(%q) expected error", tt.input)
				return
			}
			assert.NoError(t, err, "ParseMb(%q) unexpected error", tt.input)
			assert.Equal(t, tt.expected, result, "ParseMb(%q) result", tt.input)
		})
	}
}

func TestBytesToMb(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected uint64
	}{
		{
			name:     "1MiB exact",
			input:    1048576, // 1MiB in binary units
			expected: 1,       // 1 MiB
		},
		{
			name:     "1GiB",
			input:    1073741824, // 1GiB in binary units
			expected: 1024,       // 1024 MiB
		},
		{
			name:     "500KiB",
			input:    512000, // 500KiB
			expected: 1,      // Rounded up to 1 MiB
		},
		{
			name:     "15MiB",
			input:    15728640, // 15MiB
			expected: 15,       // 15 MiB
		},
		{
			name:     "25MiB",
			input:    26214400, // 25MiB
			expected: 25,       // 25 MiB
		},
		{
			name:     "100MiB exact",
			input:    104857600, // 100MiB
			expected: 100,       // 100 MiB
		},
		{
			name:     "zero bytes",
			input:    0,
			expected: 0,
		},
		{
			name:     "1 byte",
			input:    1,
			expected: 1, // Rounded up to nearest MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToMb(tt.input)
			assert.Equal(t, tt.expected, result, "BytesToMb(%d) result", tt.input)
		})
	}
}

func TestMbToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected uint64
	}{
		{
			name:     "1MiB",
			input:    1,
			expected: 1048576, // 1MiB in binary units
		},
		{
			name:     "10MiB",
			input:    10,
			expected: 10485760, // 10MiB in binary units
		},
		{
			name:     "100MiB",
			input:    100,
			expected: 104857600, // 100MiB in binary units
		},
		{
			name:     "1024MiB (1GiB)",
			input:    1024,
			expected: 1073741824, // 1GiB in binary units
		},
		{
			name:     "zero MB",
			input:    0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MbToBytes(tt.input)
			assert.Equal(t, tt.expected, result, "MbToBytes(%d) result", tt.input)
		})
	}
}

func TestFormatMb(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{
			name:     "1MiB",
			input:    1,
			expected: "1M", // Binary formatting
		},
		{
			name:     "10MiB",
			input:    10,
			expected: "10M", // Binary formatting
		},
		{
			name:     "100MiB",
			input:    100,
			expected: "100M", // Binary formatting
		},
		{
			name:     "1024MiB (1GiB)",
			input:    1024,
			expected: "1024M", // Binary formatting for 1GiB
		},
		{
			name:     "2500MB (2.5GB)",
			input:    2500,
			expected: "2500M", // Binary formatting
		},
		{
			name:     "zero MB",
			input:    0,
			expected: "0M", // Binary formatting
		},
		{
			name:     "large value - 5GB",
			input:    5000,
			expected: "5000M", // Binary formatting
		},
		{
			name:     "near uint32 overflow boundary - 4294MB",
			input:    4294,
			expected: "4294M", // Binary formatting
		},
		{
			name:     "larger value - 10000MB",
			input:    10000,
			expected: "10000M", // Binary formatting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMb(tt.input)
			assert.Equal(t, tt.expected, result, "FormatMb(%d) result", tt.input)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "1000 bytes",
			input:    1000,
			expected: "1 kB",
		},
		{
			name:     "1MB",
			input:    1000000,
			expected: "1 MB",
		},
		{
			name:     "1GB",
			input:    1000000000,
			expected: "1 GB",
		},
		{
			name:     "1TB",
			input:    1000000000000,
			expected: "1 TB",
		},
		{
			name:     "2.5GB",
			input:    2500000000,
			expected: "2.5 GB",
		},
		{
			name:     "zero bytes",
			input:    0,
			expected: "0 B",
		},
		{
			name:     "1 byte",
			input:    1,
			expected: "1 B",
		},
		{
			name:     "999 bytes",
			input:    999,
			expected: "999 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			assert.Equal(t, tt.expected, result, "FormatBytes(%d) result", tt.input)
		})
	}
}

// Test round-trip conversion
func TestRoundTripConversion(t *testing.T) {
	testMbValues := []uint64{10, 100, 1000, 2500}

	for _, mb := range testMbValues {
		t.Run("round_trip", func(t *testing.T) {
			bytes := MbToBytes(mb)
			backToMb := BytesToMb(bytes)

			// Since BytesToMb rounds up to nearest 10MB, we expect the result
			// to be either equal or the next 10MB increment
			assert.GreaterOrEqual(t, backToMb, mb, "Round trip lost precision: %d MB -> %d bytes -> %d MB", mb, bytes, backToMb)

			// The difference should be at most 10MB due to rounding
			assert.LessOrEqual(t, backToMb-mb, uint64(10), "Round trip too much rounding: %d MB -> %d bytes -> %d MB", mb, bytes, backToMb)
		})
	}
}
