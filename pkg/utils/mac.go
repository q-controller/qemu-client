package utils

import (
	"crypto/rand"
	"fmt"
)

func GenerateRandomMAC() (string, error) {
	// 6 bytes for a MAC address
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// Set locally administered bit (2nd bit of 1st byte) and clear multicast bit (1st bit)
	buf[0] = (buf[0] | 0x02) & 0xFE
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}
