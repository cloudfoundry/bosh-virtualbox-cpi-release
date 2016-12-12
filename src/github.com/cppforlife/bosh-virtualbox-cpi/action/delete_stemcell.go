package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
)

type DeleteStemcell struct {
	stemcellFinder bstem.Finder
}

func NewDeleteStemcell(stemcellFinder bstem.Finder) DeleteStemcell {
	return DeleteStemcell{stemcellFinder: stemcellFinder}
}

func (a DeleteStemcell) Run(stemcellCID StemcellCID) (interface{}, error) {
	stemcell, err := a.stemcellFinder.Find(string(stemcellCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	err = stemcell.Delete()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting stemcell '%s'", stemcellCID)
	}

	return nil, nil
}
