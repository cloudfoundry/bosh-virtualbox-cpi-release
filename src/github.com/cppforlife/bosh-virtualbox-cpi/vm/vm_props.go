package vm

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type VMProps struct {
	Memory        int
	CPUs          int
	EphemeralDisk int `json:"ephemeral_disk"`

	GUI              bool
	ParavirtProvider string `json:"paravirtprovider"`
}

func NewVMProps(props apiv1.VMCloudProps) (VMProps, error) {
	vmProps := VMProps{
		Memory:        512,
		CPUs:          1,
		EphemeralDisk: 5000,

		ParavirtProvider: "minimal", // KVM caused CPU lockups with 4+ kernel
	}

	err := props.As(&vmProps)
	if err != nil {
		return VMProps{}, err
	}

	return vmProps, nil
}
