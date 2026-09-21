package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/diskfs/go-diskfs/backend/file"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"gopkg.in/yaml.v3"
)

// CreateCloudInitISO writes user-data/meta-data/network-config under dir
// and packs them into cidata.iso.
func CreateCloudInitISO(userData, networkConfig, dir, instanceID string) (string, error) {
	userDataPath := filepath.Join(dir, "user-data")
	mergedUserData, mergeErr := mergeCloudConfig(userData)
	if mergeErr != nil {
		slog.Error("Failed to merge cloud-init config.Using the original userdata", "error", mergeErr)
		mergedUserData = userData
	}

	if err := os.WriteFile(userDataPath, []byte(mergedUserData), 0600); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, instanceID, instanceID)
	metaDataPath := filepath.Join(dir, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0600); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	networkConfigPath := filepath.Join(dir, "network-config")
	if err := os.WriteFile(networkConfigPath, []byte(networkConfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	isoPath := filepath.Join(dir, "cidata.iso")
	if isoErr := writeISO(dir, isoPath, []string{"user-data", "meta-data", "network-config"}); isoErr != nil {
		return "", isoErr
	}

	return isoPath, nil
}

// writeISO packs the named files under dir into a cidata ISO at isoPath.
// Rock Ridge keeps the names as they are for the guest kernel; Joliet does
// the same for guests that only read that.
func writeISO(dir, isoPath string, names []string) error {
	workspace, err := os.MkdirTemp("", "cidata")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(workspace) }()

	iso, err := os.Create(isoPath)
	if err != nil {
		return err
	}
	defer iso.Close()

	fs, err := iso9660.Create(file.New(iso, false), 0, 0, 0, workspace)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := addFile(fs, filepath.Join(dir, name), "/"+name); err != nil {
			return fmt.Errorf("failed to add %s: %w", name, err)
		}
	}
	return fs.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		Joliet:           true,
		VolumeIdentifier: "cidata",
	})
}

func addFile(fs *iso9660.FileSystem, src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := fs.OpenFile(dst, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func mergeCloudConfig(userdata string) (string, error) {
	var config map[string]any
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(userdata)), &config); err != nil {
		return userdata, fmt.Errorf("invalid YAML provided: %w", err)
	}

	if config == nil {
		config = make(map[string]any)
	}

	// Set resize_rootfs: true only if not present
	if _, exists := config["resize_rootfs"]; !exists {
		config["resize_rootfs"] = true
	}

	// Merge growpart only if not present
	if _, exists := config["growpart"]; !exists {
		growpart := make(map[string]any)
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
