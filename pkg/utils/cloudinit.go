package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateCloudInitISO(userData, dir, instanceID string) (string, error) {
	userDataPath := filepath.Join(dir, "user-data")
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %v", err)
	}

	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, instanceID, instanceID)
	metaDataPath := filepath.Join(dir, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %v", err)
	}

	isoPath := filepath.Join(dir, "cidata.iso")
	if isoErr := createCloudInitISOImpl(dir, isoPath); isoErr != nil {
		return "", isoErr
	}

	return isoPath, nil
}
