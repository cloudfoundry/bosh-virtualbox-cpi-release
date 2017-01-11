package vm

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (vm VMImpl) ConfigureAgent(agentEnv AgentEnv) error {
	bytes, err := json.Marshal(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	err = vm.store.Put("env.json", bytes)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env")
	}

	return nil
}

func (vm VMImpl) reconfigureAgent(hotPlug bool, agentEnvFunc func(AgentEnv) AgentEnv) error {
	prevContents, err := vm.store.Get("env.json")
	if err != nil {
		return bosherr.WrapError(err, "Fetching agent env")
	}

	var prevAgentEnv AgentEnv

	err = json.Unmarshal(prevContents, &prevAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshalling agent env")
	}

	updatedAgentEnv := agentEnvFunc(prevAgentEnv)

	err = vm.ConfigureAgent(updatedAgentEnv)
	if err != nil {
		return err
	}

	newContents, err := json.Marshal(updatedAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	err = vm.store.Put("env.json", newContents)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env")
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
