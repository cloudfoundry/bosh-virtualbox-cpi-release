package main

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	action "github.com/cppforlife/bosh-virtualbox-cpi/action"
)

type Config action.ConcreteFactoryOptions

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Reading config '%s'", path)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = action.ConcreteFactoryOptions(config).Validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating configuration")
	}

	return config, nil
}
