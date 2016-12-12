package disk

type Creator interface {
	Create(int) (Disk, error)
}

var _ Creator = Factory{}

type Finder interface {
	Find(string) (Disk, error)
}

var _ Finder = Factory{}

type Disk interface {
	ID() string

	Path() string
	VMDKPath() string

	Exists() (bool, error)
	Delete() error
}

var _ Disk = DiskImpl{}
