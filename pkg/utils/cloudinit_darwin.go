package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func createCloudInitISOImpl(cloudInitPath, isoPath string) error {
	cmd := exec.Command("mkisofs", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", fmt.Sprintf("%s/user-data", cloudInitPath), fmt.Sprintf("%s/meta-data", cloudInitPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
