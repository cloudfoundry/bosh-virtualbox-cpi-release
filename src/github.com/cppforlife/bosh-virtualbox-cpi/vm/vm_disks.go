package vm

import (
	"encoding/json"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
)

func (vm VMImpl) DiskIDs() ([]apiv1.DiskCID, error) {
	ids, err := diskAttachmentRecords{vm.store}.List()
	if err != nil {
		return nil, err
	}

	var persistentIDs []apiv1.DiskCID

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

func (vm VMImpl) AttachDisk(disk bdisk.Disk) (apiv1.DiskHint, error) {
	return vm.attachDisk(disk, false)
}

func (vm VMImpl) AttachEphemeralDisk(disk bdisk.Disk) error {
	_, err := vm.attachDisk(disk, true)
	return err
}

func (vm VMImpl) attachDisk(disk bdisk.Disk, ephemeral bool) (apiv1.DiskHint, error) {
	pd, err := vm.portDevices.FindAvailable()
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	// Actually attach the disk
	err = vm.hotPlugIfNecessary(!ephemeral, func() error { return pd.Attach(disk.VMDKPath()) })
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	rec := diskAttachmentRecord{
		ID:        disk.ID().AsString(),
		Ephemeral: ephemeral,

		Controller: pd.Controller(),
		Port:       pd.Port(),
		Device:     pd.Device(),
	}

	err = diskAttachmentRecords{vm.store}.Save(disk.ID(), rec)
	if err != nil {
		return apiv1.DiskHint{}, err
	}

	stemVer, err := vm.stemcellAPIVersion.Value()
	if err != nil {
		return apiv1.DiskHint{}, bosherr.WrapErrorf(err, "Obtaining stemcell API version")
	}

	// Update agent env for stemcells that do not support mount_diskV2
	if ephemeral || stemVer < 2 {
		vm.logger.Debug("VMImpl", "Reconfiguring agent")

		agentUpdateFunc := func(agentEnv apiv1.AgentEnv) {
			if ephemeral {
				agentEnv.AttachEphemeralDisk(pd.Hint())
			} else {
				agentEnv.AttachPersistentDisk(disk.ID(), pd.Hint())
			}
		}

		err = vm.reconfigureAgent(!ephemeral, agentUpdateFunc)
		if err != nil {
			return apiv1.DiskHint{}, bosherr.WrapErrorf(err, "Reconfiguring agent after attaching disk")
		}
	} else {
		vm.logger.Debug("VMImpl", "Skipping agent reconfiguration")
	}

	return pd.Hint(), nil
}

func (vm VMImpl) DetachDisk(disk bdisk.Disk) error {
	rec, err := diskAttachmentRecords{vm.store}.Get(disk.ID())
	if err != nil {
		return err
	}

	pd, err := vm.portDevices.Find(rec.Controller, rec.Port, rec.Device)
	if err != nil {
		return err
	}

	// Actually detach the disk
	err = vm.hotPlug(pd.Detach)
	if err != nil {
		return err
	}

	err = diskAttachmentRecords{vm.store}.Delete(disk.ID())
	if err != nil {
		return err
	}

	agentUpdateFunc := func(agentEnv apiv1.AgentEnv) {
		agentEnv.DetachPersistentDisk(disk.ID())
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

		pd, err := vm.portDevices.Find(rec.Controller, rec.Port, rec.Device)
		if err != nil {
			return err
		}

		err = pd.Detach()
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

	Controller string // e.g. scsi, ide
	Port       string // e.g. "0"
	Device     string // e.g. "1"
}

type diskAttachmentRecords struct {
	store Store
}

const (
	diskAttachmentRecordsSuffix = "-disk-attachment.json"
)

func (r diskAttachmentRecords) List() ([]apiv1.DiskCID, error) {
	keys, err := r.store.List()
	if err != nil {
		return nil, bosherr.WrapError(err, "Listing disk attachments")
	}

	var ids []apiv1.DiskCID

	for _, key := range keys {
		if !strings.HasSuffix(key, diskAttachmentRecordsSuffix) {
			continue
		}

		ids = append(ids, apiv1.NewDiskCID(strings.TrimSuffix(key, diskAttachmentRecordsSuffix)))
	}

	return ids, nil
}

func (r diskAttachmentRecords) Get(cid apiv1.DiskCID) (diskAttachmentRecord, error) {
	var rec diskAttachmentRecord

	bytes, err := r.store.Get(cid.AsString() + diskAttachmentRecordsSuffix)
	if err != nil {
		return rec, bosherr.WrapError(err, "Getting disk attachment")
	}

	err = json.Unmarshal(bytes, &rec)
	if err != nil {
		return rec, bosherr.WrapError(err, "Deserializing disk attachment")
	}

	return rec, nil
}

func (r diskAttachmentRecords) Save(cid apiv1.DiskCID, rec diskAttachmentRecord) error {
	bytes, err := json.Marshal(rec)
	if err != nil {
		return bosherr.WrapError(err, "Serializing disk attachment")
	}

	err = r.store.Put(cid.AsString()+diskAttachmentRecordsSuffix, bytes)
	if err != nil {
		return bosherr.WrapError(err, "Saving disk attachment")
	}

	return nil
}

func (r diskAttachmentRecords) Delete(cid apiv1.DiskCID) error {
	return r.store.DeleteOne(cid.AsString() + diskAttachmentRecordsSuffix)
}
