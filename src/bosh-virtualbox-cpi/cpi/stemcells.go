package cpi

import (
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bstem "bosh-virtualbox-cpi/stemcell"
)

type Stemcells struct {
	importer bstem.Importer
	finder   bstem.Finder
}

func NewStemcells(importer bstem.Importer, finder bstem.Finder) Stemcells {
	return Stemcells{importer, finder}
}

func (a Stemcells) CreateStemcell(
	imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {

	stemcell, err := a.importer.ImportFromPath(imagePath)
	if err != nil {
		return apiv1.StemcellCID{}, bosherr.WrapErrorf(err, "Importing stemcell from '%s'", imagePath)
	}

	return stemcell.ID(), nil
}

func (a Stemcells) DeleteStemcell(cid apiv1.StemcellCID) error {
	stemcell, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding stemcell '%s'", cid)
	}

	err = stemcell.Delete()
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting stemcell '%s'", cid)
	}

	return nil
}
