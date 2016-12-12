package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
)

type CreateDisk struct {
	diskCreator bdisk.Creator
}

type DiskCloudProperties map[string]interface{}

func NewCreateDisk(diskCreator bdisk.Creator) CreateDisk {
	return CreateDisk{diskCreator: diskCreator}
}

func (a CreateDisk) Run(size int, _ DiskCloudProperties, _ VMCID) (DiskCID, error) {
	disk, err := a.diskCreator.Create(size)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()), nil
}
