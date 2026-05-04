package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

const (
	qemuImg = "qemu-img"
)

type Info struct {
	VirtualSizeBytes uint64 `json:"virtual-size"`
	ActualSizeBytes  uint64 `json:"actual-size"`
}

type Image struct {
	Path string
}

func (i *Image) Info() (*Info, error) {
	if _, err := exec.LookPath(qemuImg); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", qemuImg, qemuImg)
	}

	command := exec.Command(qemuImg, "info", "--output=json", i.Path) //nolint:gosec // G204: i.Path is a caller-controlled image path
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

func (i *Image) Resize(bytes uint64) error {
	if _, err := exec.LookPath(qemuImg); err != nil {
		return fmt.Errorf("%s is not available; please install %s", qemuImg, qemuImg)
	}

	command := exec.Command(qemuImg, "resize", i.Path, strconv.FormatUint(bytes, 10)) //nolint:gosec // G204: i.Path is a caller-controlled image path
	_, outErr := command.Output()
	if outErr != nil {
		return outErr
	}

	return nil
}
