package vm

import (
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"

	bdisk "bosh-virtualbox-cpi/disk"
	bstem "bosh-virtualbox-cpi/stemcell"
)

type Creator interface {
	Create(
		apiv1.AgentID,
		bstem.Stemcell,
		apiv1.VMCloudProps,
		apiv1.Networks,
		apiv1.VMEnv,
	) (VM, error)
}

var _ Creator = Factory{}

type Finder interface {
	Find(apiv1.VMCID) (VM, error)
}

var _ Finder = Factory{}

type VM interface {
	ID() apiv1.VMCID
	SetMetadata(apiv1.VMMeta) error

	Reboot() error
	Exists() (bool, error)
	Delete() error

	DiskIDs() ([]apiv1.DiskCID, error)
	AttachDisk(bdisk.Disk) (apiv1.DiskHint, error)
	AttachEphemeralDisk(bdisk.Disk) error
	DetachDisk(bdisk.Disk) error
}

var _ VM = VMImpl{}
