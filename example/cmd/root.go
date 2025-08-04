package cmd

import (
	"log/slog"
	"os"

	"github.com/q-controller/qemu-client/pkg/qemu"
	"github.com/q-controller/qemu-client/pkg/utils"
	"github.com/spf13/cobra"
)

var image string
var rootCmd = &cobra.Command{
	Use:   "example",
	Short: "Example app to start qemu VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		mac, macErr := utils.GenerateRandomMAC()
		if macErr != nil {
			return macErr
		}

		instance, instanceErr := qemu.Start("example", image, qemu.Config{
			Cpus:   1,
			Memory: "1G",
			Disk:   "10G",
			HwAddr: mac,
			UserData: `#cloud-config
ssh_pwauth: true
users:
  - name: exampleuser
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    groups: sudo
    lock_passwd: false
    ssh-authorized-keys: []
chpasswd:
  list: |
    exampleuser:examplepass
  expire: false
`,
		})

		if instanceErr != nil {
			return instanceErr
		}

	channelLoop:
		for {
			select {
			case str := <-instance.Stderr:
				slog.Error(str)
			case str := <-instance.Stdout:
				slog.Info(str)
			case <-instance.Done:
				break channelLoop
			}
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&image, "image", "", "Path to the raw image")
	rootCmd.MarkFlagRequired("image")
}
