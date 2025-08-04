package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func createCloudInitISOImpl(cloudInitPath, isoPath string) error {
	cmd := exec.Command("genisoimage", "-output", isoPath, "-V", "cidata", "-r", "-J", fmt.Sprintf("%s/user-data", cloudInitPath), fmt.Sprintf("%s/meta-data", cloudInitPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd.Wait()

	return nil
}
