package stemcell

type Importer interface {
	ImportFromPath(string) (Stemcell, error)
}

var _ Importer = Factory{}

type Finder interface {
	Find(string) (Stemcell, error)
}

var _ Finder = Factory{}

type Stemcell interface {
	ID() string
	SnapshotName() string

	Exists() (bool, error)
	Delete() error
}

var _ Stemcell = StemcellImpl{}
