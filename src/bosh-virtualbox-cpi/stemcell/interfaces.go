package stemcell

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type Importer interface {
	ImportFromPath(string) (Stemcell, error)
}

var _ Importer = Factory{}

type Finder interface {
	Find(apiv1.StemcellCID) (Stemcell, error)
}

var _ Finder = Factory{}

type Stemcell interface {
	ID() apiv1.StemcellCID
	SnapshotName() string

	Exists() (bool, error)
	Delete() error
}

var _ Stemcell = StemcellImpl{}
