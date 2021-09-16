package cpi

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bpds "bosh-virtualbox-cpi/vm/portdevices"
)

type FactoryOpts struct {
	Host       string
	Username   string
	PrivateKey string

	BinPath  string
	StoreDir string

	StorageController  string
	AutoEnableNetworks bool

	Agent apiv1.AgentOptions
}

func (o FactoryOpts) Validate() error {
	if len(o.Host) > 0 {
		if o.Username == "" {
			return bosherr.Error("Must provide non-empty Username")
		}

		if o.PrivateKey == "" {
			return bosherr.Error("Must provide non-empty PrivateKey")
		}
	}

	if o.BinPath == "" {
		return bosherr.Error("Must provide non-empty BinPath")
	}

	if o.StoreDir == "" {
		return bosherr.Error("Must provide non-empty StoreDir")
	}

	switch o.StorageController {
	case bpds.IDEController, bpds.SCSIController:
		// valid
	default:
		return bosherr.Error("Unexpected StorageController")
	}

	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}

func (o FactoryOpts) StemcellsDir() string {
	return filepath.Join(o.StoreDir, "stemcells")
}

func (o FactoryOpts) VMsDir() string {
	return filepath.Join(o.StoreDir, "vms")
}

func (o FactoryOpts) DisksDir() string {
	return filepath.Join(o.StoreDir, "disks")
}
