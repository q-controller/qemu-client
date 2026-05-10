package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

const defaultQemuImg = "qemu-img"

type Info struct {
	VirtualSizeBytes uint64 `json:"virtual-size"`
	ActualSizeBytes  uint64 `json:"actual-size"`
}

// ImageTool runs qemu-img against a disk image. The disk image is the data
// (passed per-call); ImageTool is the operator (constructed once with the
// resolved binary path).
type ImageTool struct {
	QemuImg string // absolute path; empty = PATH lookup of "qemu-img"
}

func (t ImageTool) binary() string {
	if t.QemuImg != "" {
		return t.QemuImg
	}
	return defaultQemuImg
}

func (t ImageTool) Info(path string) (*Info, error) {
	bin := t.binary()
	if _, err := exec.LookPath(bin); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", bin, bin)
	}

	command := exec.Command(bin, "info", "--output=json", path) //nolint:gosec // G204: path is caller-controlled
	bytes, bytesErr := command.Output()
	if bytesErr != nil {
		return nil, bytesErr
	}

	var info Info
	if unmarshalErr := json.Unmarshal(bytes, &info); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &info, nil
}

func (t ImageTool) Resize(path string, sizeBytes uint64) error {
	bin := t.binary()
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%s is not available; please install %s", bin, bin)
	}

	command := exec.Command(bin, "resize", path, strconv.FormatUint(sizeBytes, 10)) //nolint:gosec // G204: path is caller-controlled
	_, outErr := command.Output()
	if outErr != nil {
		return outErr
	}

	return nil
}
