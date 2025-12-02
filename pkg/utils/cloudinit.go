package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func CreateCloudInitISO(userData, dir, instanceID string) (string, error) {
	userDataPath := filepath.Join(dir, "user-data")
	mergedUserData, mergeErr := mergeCloudConfig(userData)
	if mergeErr != nil {
		fmt.Printf("Failed to merge cloud-init config: %v.\nUsing the original userdata.\n", mergeErr)
		mergedUserData = userData
	}

	if err := os.WriteFile(userDataPath, []byte(mergedUserData), 0644); err != nil {
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

func mergeCloudConfig(userdata string) (string, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(userdata)), &config); err != nil {
		return userdata, fmt.Errorf("invalid YAML provided: %v", err)
	}

	if config == nil {
		config = make(map[string]interface{})
	}

	// Set resize_rootfs: true only if not present
	if _, exists := config["resize_rootfs"]; !exists {
		config["resize_rootfs"] = true
	}

	// Merge growpart only if not present
	if _, exists := config["growpart"]; !exists {
		growpart := make(map[string]interface{})
		growpart["mode"] = "auto"
		growpart["devices"] = []string{"/"}
		config["growpart"] = growpart
	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}

	return "#cloud-config\n" + string(out), nil
}
