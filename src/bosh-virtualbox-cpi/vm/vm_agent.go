package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"
)

func (vm VMImpl) ConfigureAgent(agentEnv apiv1.AgentEnv) error {
	_, err := vm.configureAgent(agentEnv)
	return err
}

func (vm VMImpl) configureAgent(agentEnv apiv1.AgentEnv) ([]byte, error) {
	bytes, err := agentEnv.AsBytes()
	if err != nil {
		return nil, bosherr.WrapError(err, "Marshalling agent env")
	}

	err = vm.store.Put("env.json", bytes)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating agent env")
	}

	return bytes, nil
}

func (vm VMImpl) reconfigureAgent(hotPlug bool, agentEnvFunc func(apiv1.AgentEnv)) error {
	// todo hide unmarshaling within apiv1
	prevContents, err := vm.store.Get("env.json")
	if err != nil {
		return bosherr.WrapError(err, "Fetching agent env")
	}

	agentEnv, err := apiv1.NewAgentEnvFactory().FromBytes(prevContents)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshalling agent env")
	}

	agentEnvFunc(agentEnv)

	newContents, err := vm.configureAgent(agentEnv)
	if err != nil {
		return err
	}

	isoBytes, err := ISO9660{FileName: "ENV", Contents: newContents}.Bytes()
	if err != nil {
		return bosherr.WrapError(err, "Marshaling agent env to ISO")
	}

	err = vm.store.Put("env.iso", isoBytes)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env")
	}

	updateFunc := func() error {
		cd, err := vm.portDevices.CDROM()
		if err != nil {
			return err
		}

		return cd.Mount(vm.store.Path("env.iso"))
	}

	return vm.hotPlugIfNecessary(hotPlug, updateFunc)
}
