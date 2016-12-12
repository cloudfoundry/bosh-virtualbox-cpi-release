package vm

import (
	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
)

type Creator interface {
	Create(string, bstem.Stemcell, VMProps, Networks, Environment) (VM, error)
}

var _ Creator = Factory{}

type Finder interface {
	Find(string) (VM, error)
}

var _ Finder = Factory{}

type VMProps struct {
	Memory        int // 512
	CPUs          int // 1
	EphemeralDisk int

	GUI              bool
	ParavirtProvider string // minimal
}

type VMMetadata map[string]interface{}

type Environment map[string]interface{}

type VM interface {
	ID() string
	SetMetadata(VMMetadata) error

	Reboot() error
	Exists() (bool, error)
	Delete() error

	DiskIDs() ([]string, error)
	AttachDisk(bdisk.Disk) error
	AttachEphemeralDisk(bdisk.Disk) error
	DetachDisk(bdisk.Disk) error
}

var _ VM = VMImpl{}
