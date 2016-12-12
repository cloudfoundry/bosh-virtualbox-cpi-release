package vm

import (
	"encoding/json"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
)

func (vm VMImpl) DiskIDs() ([]string, error) {
	ids, err := diskAttachmentRecords{vm.store}.List()
	if err != nil {
		return nil, err
	}

	var persistentIDs []string

	for _, id := range ids {
		rec, err := diskAttachmentRecords{vm.store}.Get(id)
		if err != nil {
			return nil, err
		} else if !rec.Ephemeral {
			persistentIDs = append(persistentIDs, id)
		}
	}

	return persistentIDs, nil
}

func (vm VMImpl) AttachDisk(disk bdisk.Disk) error {
	return vm.attachDisk(disk, false)
}

func (vm VMImpl) AttachEphemeralDisk(disk bdisk.Disk) error {
	return vm.attachDisk(disk, true)
}

func (vm VMImpl) attachDisk(disk bdisk.Disk, ephemeral bool) error {
	pd, err := PortDevices{vm.driver, vm}.FindAvailable()
	if err != nil {
		return err
	}

	// Actually attach the disk
	err = vm.hotPlugIfNecessary(!ephemeral, func() error { return pd.Attach(disk.VMDKPath()) })
	if err != nil {
		return err
	}

	rec := diskAttachmentRecord{
		ID:        disk.ID(),
		Ephemeral: ephemeral,

		Port:   pd.port,
		Device: pd.device,
	}

	err = diskAttachmentRecords{vm.store}.Save(disk.ID(), rec)
	if err != nil {
		return err
	}

	agentUpdateFunc := func(agentEnv AgentEnv) AgentEnv {
		if ephemeral {
			return agentEnv.AttachEphemeralDisk(pd.port)
		} else {
			return agentEnv.AttachPersistentDisk(disk.ID(), pd.port)
		}
	}

	err = vm.reconfigureAgent(!ephemeral, agentUpdateFunc)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reconfiguring agent after attaching disk")
	}

	return nil
}

func (vm VMImpl) DetachDisk(disk bdisk.Disk) error {
	rec, err := diskAttachmentRecords{vm.store}.Get(disk.ID())
	if err != nil {
		return err
	}

	// Actually detach the disk
	err = vm.hotPlug(PortDevices{vm.driver, vm}.Find(rec.Port, rec.Device).Detach)
	if err != nil {
		return err
	}

	err = diskAttachmentRecords{vm.store}.Delete(disk.ID())
	if err != nil {
		return err
	}

	agentUpdateFunc := func(agentEnv AgentEnv) AgentEnv {
		return agentEnv.DetachPersistentDisk(disk.ID())
	}

	err = vm.reconfigureAgent(false, agentUpdateFunc)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reconfiguring agent after detaching disk")
	}

	return nil
}

func (vm VMImpl) detachPersistentDisks() error {
	recs := diskAttachmentRecords{vm.store}

	ids, err := recs.List()
	if err != nil {
		return err
	}

	for _, id := range ids {
		rec, err := recs.Get(id)
		if err != nil {
			return err
		} else if rec.Ephemeral {
			continue
		}

		err = PortDevices{vm.driver, vm}.Find(rec.Port, rec.Device).Detach()
		if err != nil {
			return err
		}

		err = recs.Delete(id)
		if err != nil {
			return err
		}
	}

	return nil
}

type diskAttachmentRecord struct {
	ID        string
	Ephemeral bool

	Port   string
	Device string
}

type diskAttachmentRecords struct {
	store Store
}

const (
	diskAttachmentRecordsSuffix = "-disk-attachment.json"
)

func (r diskAttachmentRecords) List() ([]string, error) {
	keys, err := r.store.List()
	if err != nil {
		return nil, bosherr.WrapError(err, "Listing disk attachments")
	}

	var ids []string

	for _, key := range keys {
		if !strings.HasSuffix(key, diskAttachmentRecordsSuffix) {
			continue
		}

		ids = append(ids, strings.TrimSuffix(key, diskAttachmentRecordsSuffix))
	}

	return ids, nil
}

func (r diskAttachmentRecords) Get(diskID string) (diskAttachmentRecord, error) {
	var rec diskAttachmentRecord

	bytes, err := r.store.Get(diskID + diskAttachmentRecordsSuffix)
	if err != nil {
		return rec, bosherr.WrapError(err, "Getting disk attachment")
	}

	err = json.Unmarshal(bytes, &rec)
	if err != nil {
		return rec, bosherr.WrapError(err, "Deserializing disk attachment")
	}

	return rec, nil
}

func (r diskAttachmentRecords) Save(diskID string, rec diskAttachmentRecord) error {
	bytes, err := json.Marshal(rec)
	if err != nil {
		return bosherr.WrapError(err, "Serializing disk attachment")
	}

	err = r.store.Put(diskID+diskAttachmentRecordsSuffix, bytes)
	if err != nil {
		return bosherr.WrapError(err, "Saving disk attachment")
	}

	return nil
}

func (r diskAttachmentRecords) Delete(diskID string) error {
	return r.store.DeleteOne(diskID + diskAttachmentRecordsSuffix)
}
