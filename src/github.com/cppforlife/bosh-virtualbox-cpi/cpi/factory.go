package cpi

import (
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type Factory struct {
	fs         boshsys.FileSystem
	cmdRunner  boshsys.CmdRunner
	uuidGen    boshuuid.Generator
	compressor boshcmd.Compressor
	opts       FactoryOpts
	logger     boshlog.Logger
}

var _ apiv1.CPIFactory = Factory{}

type CPI struct {
	Misc
	Stemcells
	VMs
	Disks
	Snapshots
}

var _ apiv1.CPI = CPI{}

func NewFactory(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	compressor boshcmd.Compressor,
	opts FactoryOpts,
	logger boshlog.Logger,
) Factory {
	return Factory{fs, cmdRunner, uuidGen, compressor, opts, logger}
}

func (f Factory) New(ctx apiv1.CallContext) (apiv1.CPI, error) {
	retrier := driver.RetrierImpl{}
	rawRunner := driver.RawRunner(driver.NewLocalRunner(f.fs, f.cmdRunner, f.logger))

	if len(f.opts.Host) > 0 {
		runnerOpts := driver.SSHRunnerOpts{
			Host:       f.opts.Host,
			Username:   f.opts.Username,
			PrivateKey: f.opts.PrivateKey,
		}
		rawRunner = driver.NewSSHRunner(runnerOpts, f.fs, f.logger)
	}

	runner := driver.NewExpandingPathRunner(rawRunner)
	driver := driver.NewExecDriver(runner, retrier, f.opts.BinPath, f.logger)

	stemcellsOpts := bstem.FactoryOpts{
		DirPath:           f.opts.StemcellsDir(),
		StorageController: f.opts.StorageController,
	}

	stemcells := bstem.NewFactory(
		stemcellsOpts, driver, runner, retrier, f.fs, f.uuidGen, f.compressor, f.logger)

	disks := bdisk.NewFactory(f.opts.DisksDir(), f.uuidGen, driver, runner, f.logger)

	vmsOpts := bvm.FactoryOpts{
		DirPath:            f.opts.VMsDir(),
		StorageController:  f.opts.StorageController,
		AutoEnableNetworks: f.opts.AutoEnableNetworks,
	}

	vms := bvm.NewFactory(
		vmsOpts, f.uuidGen, driver, runner, disks,
		f.opts.Agent, apiv1.NewStemcellAPIVersion(ctx), f.logger)

	return CPI{
		NewMisc(),
		NewStemcells(stemcells, stemcells),
		NewVMs(stemcells, vms, vms),
		NewDisks(disks, disks, vms),
		NewSnapshots(),
	}, nil
}
