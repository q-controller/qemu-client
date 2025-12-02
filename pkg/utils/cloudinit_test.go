package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// parseCloudConfig removes the #cloud-config header and parses the YAML content
func parseCloudConfig(t *testing.T, content string) map[string]interface{} {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#cloud-config") {
		lines = lines[1:]
	}
	yamlContent := strings.Join(lines, "\n")

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	require.NoError(t, err, "Failed to parse YAML")
	return config
}

func TestMergeCloudConfig_EmptyUserdata(t *testing.T) {
	userdata := ""

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)
	assert.Equal(t, true, config["resize_rootfs"])

	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "auto", growpart["mode"])
	devices, ok := growpart["devices"].([]interface{})
	require.True(t, ok, "devices should be an array")
	assert.Equal(t, "/", devices[0])
}

func TestMergeCloudConfig_InvalidYAML(t *testing.T) {
	userdata := "invalid: yaml: :"

	result, err := mergeCloudConfig(userdata)
	require.Error(t, err)

	assert.Equal(t, result, "invalid: yaml: :")
}

func TestMergeCloudConfig_ValidYAMLWithoutHeader(t *testing.T) {
	userdata := `users:
  - name: testuser
    shell: /bin/bash`

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)

	// Should preserve user configuration
	users, ok := config["users"].([]interface{})
	require.True(t, ok, "users should be an array")
	require.Len(t, users, 1)
	user, ok := users[0].(map[string]interface{})
	require.True(t, ok, "user should be a map")
	assert.Equal(t, "testuser", user["name"])
	assert.Equal(t, "/bin/bash", user["shell"])

	// Should add resize configuration
	assert.Equal(t, true, config["resize_rootfs"])
	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "auto", growpart["mode"])
}

func TestMergeCloudConfig_ExistingGrowpartPreserved(t *testing.T) {
	userdata := `#cloud-config
resize_rootfs: false
growpart:
  mode: manual
  devices: ['/dev/sda1']`

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)

	assert.Equal(t, false, config["resize_rootfs"])

	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "manual", growpart["mode"])
	devices, ok := growpart["devices"].([]interface{})
	require.True(t, ok, "devices should be an array")
	assert.Equal(t, "/dev/sda1", devices[0])
}

func TestMergeCloudConfig_PreserveExistingFields(t *testing.T) {
	userdata := `#cloud-config
users:
  - name: testuser
    sudo: ALL=(ALL) NOPASSWD:ALL
packages:
  - git
  - vim`

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)

	users, ok := config["users"].([]interface{})
	require.True(t, ok, "users should be an array")
	require.Len(t, users, 1)
	user, ok := users[0].(map[string]interface{})
	require.True(t, ok, "user should be a map")
	assert.Equal(t, "testuser", user["name"])
	assert.Equal(t, "ALL=(ALL) NOPASSWD:ALL", user["sudo"])

	packages, ok := config["packages"].([]interface{})
	require.True(t, ok, "packages should be an array")
	assert.Contains(t, packages, "git")
	assert.Contains(t, packages, "vim")

	assert.Equal(t, true, config["resize_rootfs"])
	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "auto", growpart["mode"])
}

func TestMergeCloudConfig_WhitespaceOnly(t *testing.T) {
	userdata := "   \n\t  \n   "

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)
	assert.Equal(t, true, config["resize_rootfs"])
	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "auto", growpart["mode"])
}

func TestMergeCloudConfig_ComplexConfiguration(t *testing.T) {
	userdata := `#cloud-config
users:
  - name: ubuntu
    ssh_authorized_keys:
      - ssh-rsa AAAAB3...
runcmd:
  - echo 'Hello World'
package_update: true
resize_rootfs: false`

	result, err := mergeCloudConfig(userdata)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "#cloud-config"))

	config := parseCloudConfig(t, result)

	users, ok := config["users"].([]interface{})
	require.True(t, ok, "users should be an array")
	require.Len(t, users, 1)
	user, ok := users[0].(map[string]interface{})
	require.True(t, ok, "user should be a map")
	assert.Equal(t, "ubuntu", user["name"])

	keys, ok := user["ssh_authorized_keys"].([]interface{})
	require.True(t, ok, "ssh_authorized_keys should be an array")
	assert.Contains(t, keys, "ssh-rsa AAAAB3...")

	runcmd, ok := config["runcmd"].([]interface{})
	require.True(t, ok, "runcmd should be an array")
	assert.Contains(t, runcmd, "echo 'Hello World'")

	assert.Equal(t, true, config["package_update"])
	assert.Equal(t, false, config["resize_rootfs"])

	// Should add growpart only if missing
	growpart, ok := config["growpart"].(map[string]interface{})
	require.True(t, ok, "growpart should be a map")
	assert.Equal(t, "auto", growpart["mode"])
}
