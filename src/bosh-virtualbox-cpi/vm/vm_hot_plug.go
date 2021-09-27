package vm

func (vm VMImpl) hotPlugIfNecessary(necessary bool, innerFunc func() error) error {
	if necessary {
		return vm.hotPlug(innerFunc)
	}
	return innerFunc()
}

func (vm VMImpl) hotPlug(innerFunc func() error) error {
	// VirtualBox as of 4.2.16 does not support real device hot plugging.
	// (http://www.youtube.com/watch?v=5c1m2BAg2Sc)

	// http://dlc.sun.com.edgesuite.net/virtualbox/4.2.16/UserManual.pdf
	// Section 9.25: VirtualBox expert storage management
	_, err := vm.driver.Execute("setextradata", vm.cid.AsString(), "VBoxInternal2/SilentReconfigureWhilePaused", "1")
	if err != nil {
		return err
	}

	var needsToResume bool

	running, err := vm.IsRunning()
	if err != nil {
		return err
	}

	if running {
		needsToResume = true
		_, err = vm.driver.Execute("controlvm", vm.cid.AsString(), "pause")
		if err != nil {
			return err
		}
	}

	err = innerFunc()
	if err != nil {
		return err
	}

	// todo should always do this?
	if needsToResume {
		_, err = vm.driver.Execute("controlvm", vm.cid.AsString(), "resume")
	}

	return err
}
