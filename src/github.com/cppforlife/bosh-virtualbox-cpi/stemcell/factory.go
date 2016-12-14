package stemcell

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	stemcellSuggestedName = regexp.MustCompile(`Suggested VM name "(.+?)"`)
)

type Factory struct {
	dirPath string

	driver  driver.Driver
	runner  driver.Runner
	retrier driver.Retrier

	fs         boshsys.FileSystem
	uuidGen    boshuuid.Generator
	compressor boshcmd.Compressor

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	dirPath string,
	driver driver.Driver,
	runner driver.Runner,
	retrier driver.Retrier,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	compressor boshcmd.Compressor,
	logger boshlog.Logger,
) Factory {
	return Factory{
		dirPath: dirPath,

		driver:  driver,
		runner:  runner,
		retrier: retrier,

		fs:         fs,
		uuidGen:    uuidGen,
		compressor: compressor,

		logTag: "stemcell.Factory",
		logger: logger,
	}
}

func (f Factory) ImportFromPath(imagePath string) (Stemcell, error) {
	id, err := f.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating stemcell id")
	}

	id = "sc-" + id

	stemcellPath := filepath.Join(f.dirPath, id)

	err = f.upload(imagePath, stemcellPath)
	if err != nil {
		return nil, err
	}

	internalTmpID, err := f.importOVF(filepath.Join(stemcellPath, "image.ovf"))
	if err != nil {
		return nil, err
	}

	// todo remove stemcell after import?

	_, err = f.driver.Execute("modifyvm", internalTmpID, "--name", id)
	if err != nil {
		f.cleanUpPartialImport(internalTmpID)
		return nil, bosherr.WrapErrorf(err, "Setting stemcell name")
	}

	stemcell := f.newStemcell(id)

	err = stemcell.Prepare()
	if err != nil {
		f.cleanUpPartialImport(internalTmpID)
		return nil, bosherr.WrapErrorf(err, "Preparing stemcell")
	}

	return stemcell, err
}

func (f Factory) Find(id string) (Stemcell, error) {
	return f.newStemcell(id), nil
}

func (f Factory) newStemcell(id string) StemcellImpl {
	return NewStemcellImpl(id, filepath.Join(f.dirPath, id), f.driver, f.runner, f.logger)
}

func (f Factory) upload(imagePath, stemcellPath string) error {
	tmpDir, err := f.fs.TempDir("bosh-virtualbox-cpi-stemcell-upload")
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating tmp stemcell directory")
	}

	defer f.fs.RemoveAll(tmpDir)

	err = f.compressor.DecompressFileToDir(imagePath, tmpDir, boshcmd.CompressorOptions{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Unpacking stemcell '%s' to '%s'", imagePath, tmpDir)
	}

	_, _, err = f.runner.Execute("mkdir", "-p", stemcellPath)
	if err != nil {
		return bosherr.WrapError(err, "Creating stemcell parent")
	}

	for _, fileName := range []string{"image-disk1.vmdk", "image.mf", "image.ovf"} {
		err := f.runner.Upload(filepath.Join(tmpDir, fileName), filepath.Join(stemcellPath, fileName))
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading stemcell")
		}
	}

	return nil
}

func (f Factory) importOVF(ovfPath string) (string, error) {
	var internalTmpID string

	actionFunc := func() error {
		output, err := f.driver.Execute("import", ovfPath)
		if err != nil {
			return driver.RetryableErrorImpl{err}
		}

		matches := stemcellSuggestedName.FindStringSubmatch(output)
		if len(matches) != 2 {
			return driver.RetryableErrorImpl{bosherr.Errorf("Couldn't find VM name in the output:\nOutput: '%s'", output)}
		}

		suggestedName := matches[1]

		output, err = f.driver.Execute("list", "vms")
		if err != nil {
			f.cleanUpPartialImport(suggestedName)
			return driver.RetryableErrorImpl{bosherr.WrapError(err, "Listing VMs after an import")}
		}

		// todo regexp.MustCompile(`^"#{Regexp.escape(suggestedName)}" \{(.+?)\}$`)

		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, fmt.Sprintf(`"%s" `, suggestedName)) {
				internalTmpID = strings.TrimSuffix(strings.SplitN(line, " {", 2)[1], "}")
				return nil
			}
		}

		f.cleanUpPartialImport(suggestedName)
		return driver.RetryableErrorImpl{bosherr.Errorf("Failed to import '%s'", ovfPath)}
	}

	return internalTmpID, f.retrier.RetryComplex(actionFunc, 2, 2*time.Second)
}

func (f Factory) cleanUpPartialImport(suggestedNameOrID string) {
	_, err := f.driver.Execute("unregistervm", suggestedNameOrID, "--delete")
	if err != nil {
		f.logger.Error(f.logTag, "Failed to clean up partially imported stemcell: %s", err)
	}
}
