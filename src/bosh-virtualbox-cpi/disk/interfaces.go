package disk

import (
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"
)

type Creator interface {
	Create(int) (Disk, error)
}

var _ Creator = Factory{}

type Finder interface {
	Find(apiv1.DiskCID) (Disk, error)
}

var _ Finder = Factory{}

type Disk interface {
	ID() apiv1.DiskCID

	Path() string
	VMDKPath() string

	Exists() (bool, error)
	Delete() error
}

var _ Disk = DiskImpl{}
