package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

const (
	qemu_img = "qemu-img"
)

type Info struct {
	VirtualSizeBytes uint64 `json:"virtual-size"`
	ActualSizeBytes  uint64 `json:"actual-size"`
}

type Image struct {
	Path string
}

func (i *Image) Info() (*Info, error) {
	if _, err := exec.LookPath(qemu_img); err != nil {
		return nil, fmt.Errorf("%s is not available; please install %s", qemu_img, qemu_img)
	}

	command := exec.Command(qemu_img, "info", "--output=json", i.Path)
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
	if _, err := exec.LookPath(qemu_img); err != nil {
		return fmt.Errorf("%s is not available; please install %s", qemu_img, qemu_img)
	}

	command := exec.Command(qemu_img, "resize", i.Path, fmt.Sprintf("%d", bytes))
	_, outErr := command.Output()
	if outErr != nil {
		return outErr
	}

	return nil
}
