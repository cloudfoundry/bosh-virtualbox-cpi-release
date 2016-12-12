package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	compressor boshcmd.Compressor,
	options ConcreteFactoryOptions,
	logger boshlog.Logger,
) concreteFactory {
	retrier := driver.RetrierImpl{}
	runner := driver.NewLocalRunner(fs, cmdRunner)
	driver := driver.NewExecDriver(runner, retrier, options.BinPath, logger)
	stemcells := bstem.NewFactory(options.StemcellsDir(), driver, runner, retrier, fs, uuidGen, compressor, logger)
	disks := bdisk.NewFactory(options.DisksDir(), uuidGen, driver, runner, logger)
	vms := bvm.NewFactory(options.VMsDir(), uuidGen, driver, runner, disks, options.Agent, logger)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcells),
			"delete_stemcell": NewDeleteStemcell(stemcells),

			// VM management
			"create_vm":       NewCreateVM(stemcells, vms),
			"delete_vm":       NewDeleteVM(vms),
			"has_vm":          NewHasVM(vms),
			"reboot_vm":       NewRebootVM(vms),
			"set_vm_metadata": NewSetVMMetadata(vms),

			// Disk management
			"create_disk": NewCreateDisk(disks),
			"delete_disk": NewDeleteDisk(disks),
			"attach_disk": NewAttachDisk(vms, disks),
			"detach_disk": NewDetachDisk(vms, disks),
			"has_disk":    NewHasDisk(disks),
			"get_disks":   NewGetDisks(vms),
		},
	}
}

func (f concreteFactory) Create(method string) (Action, error) {
	action, found := f.availableActions[method]
	if !found {
		return nil, bosherr.Errorf("Could not create action with method '%s'", method)
	}

	return action, nil
}
