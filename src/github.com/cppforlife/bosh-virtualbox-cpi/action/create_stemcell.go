package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
)

type CreateStemcell struct {
	stemcellImporter bstem.Importer
}

type CreateStemcellCloudProps struct{}

func NewCreateStemcell(stemcellImporter bstem.Importer) CreateStemcell {
	return CreateStemcell{stemcellImporter: stemcellImporter}
}

func (a CreateStemcell) Run(imagePath string, _ CreateStemcellCloudProps) (StemcellCID, error) {
	stemcell, err := a.stemcellImporter.ImportFromPath(imagePath)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Importing stemcell from '%s'", imagePath)
	}

	return StemcellCID(stemcell.ID()), nil
}
