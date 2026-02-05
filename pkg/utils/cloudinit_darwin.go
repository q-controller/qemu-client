package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func createCloudInitISOImpl(cloudInitPath, isoPath string) error {
	cmd := exec.Command("mkisofs", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock",
		fmt.Sprintf("%s/user-data", cloudInitPath),
		fmt.Sprintf("%s/meta-data", cloudInitPath),
		fmt.Sprintf("%s/network-config", cloudInitPath))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mkisofs failed: %w: %s", err, stderr.String())
	}
	return nil
}
