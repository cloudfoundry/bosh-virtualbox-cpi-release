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
	rawRunner := driver.RawRunner(driver.NewLocalRunner(fs, cmdRunner, logger))

	if len(options.Host) > 0 {
		runnerOpts := driver.SSHRunnerOpts{
			Host:       options.Host,
			Username:   options.Username,
			PrivateKey: options.PrivateKey,
		}
		rawRunner = driver.NewSSHRunner(runnerOpts, fs, logger)
	}

	runner := driver.NewExpandingPathRunner(rawRunner)
	driver := driver.NewExecDriver(runner, retrier, options.BinPath, logger)

	stemcellsOpts := bstem.FactoryOpts{
		DirPath:           options.StemcellsDir(),
		StorageController: options.StorageController,
	}

	stemcells := bstem.NewFactory(stemcellsOpts, driver, runner, retrier, fs, uuidGen, compressor, logger)
	disks := bdisk.NewFactory(options.DisksDir(), uuidGen, driver, runner, logger)

	vmsOpts := bvm.FactoryOpts{
		DirPath:            options.VMsDir(),
		StorageController:  options.StorageController,
		AutoEnableNetworks: options.AutoEnableNetworks,
	}

	vms := bvm.NewFactory(vmsOpts, uuidGen, driver, runner, disks, options.Agent, logger)

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
