package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func createCloudInitISOImpl(cloudInitPath, isoPath string) error {
	//nolint:gosec // G204: building cloud-init ISO from caller-controlled paths
	cmd := exec.Command("genisoimage", "-output", isoPath, "-V", "cidata", "-r", "-J", cloudInitPath+"/user-data", cloudInitPath+"/meta-data", cloudInitPath+"/network-config")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("genisoimage failed: %w: %s", err, stderr.String())
	}
	return nil
}
