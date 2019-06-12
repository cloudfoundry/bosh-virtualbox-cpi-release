package vm

import (
	"errors"

	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type SharedFolder struct {
	HostPath string `json:"host_path"`
}

type VMProps struct {
	Memory        int
	CPUs          int
	EphemeralDisk int    `json:"ephemeral_disk"`
	Audio         string `json:"audio"`

	GUI              bool
	ParavirtProvider string `json:"paravirtprovider"`

	SharedFolders []SharedFolder `json:"shared_folders"`
}

func NewVMProps(props apiv1.VMCloudProps) (VMProps, error) {
	vmProps := VMProps{
		Memory:        512,
		CPUs:          1,
		EphemeralDisk: 5000,
		Audio:         "none",

		ParavirtProvider: "default", // Let vboxmanage decide which paravirtprovider to use
	}

	err := props.As(&vmProps)
	if err != nil {
		return VMProps{}, err
	}

	for _, folder := range vmProps.SharedFolders {
		if folder.HostPath == "" {
			return VMProps{}, errors.New("Expected host paths not to be empty")
		}
	}

	return vmProps, nil
}
