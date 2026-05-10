package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func createCloudInitISOImpl(cloudInitPath, isoPath, isoCreator string) error {
	if isoCreator == "" {
		isoCreator = "genisoimage"
	}
	//nolint:gosec // G204: building cloud-init ISO from caller-controlled paths
	cmd := exec.Command(isoCreator, "-output", isoPath, "-V", "cidata", "-r", "-J", cloudInitPath+"/user-data", cloudInitPath+"/meta-data", cloudInitPath+"/network-config")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w: %s", isoCreator, err, stderr.String())
	}
	return nil
}
