package network

import (
	"bosh-virtualbox-cpi/driver"
	"fmt"
	"regexp"
)

var (
	createdHostOnlyMatch = regexp.MustCompile(`Interface '(.+)' was successfully created`)
)

func (n Networks) AddHostOnly(name, gateway, netmask string) (bool, error) {
	// VB does not allow naming host-only networks, exit if it's not the first one
	// if len(name) > 0 && name != "vboxnet0" {
	// 	return false, fmt.Errorf("VBoxNetwork does not have expected name, expected vboxnet0, received %s", name)
	// }

	createdName, err := n.createHostOnly()
	if err != nil {
		return true, err
	}

	if len(name) > 0 && createdName != name {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return true, fmt.Errorf("Expected created host-only network '%s' to have name '%s'", createdName, name)
	}

	err = n.configureHostOnly(createdName, gateway, netmask)
	if err != nil {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return true, err
	}

	return true, nil
}

func (n Networks) createHostOnly() (string, error) {
	output, err := n.driver.Execute("hostonlyif", "create")
	if err != nil {
		return "", err
	}

	matches := createdHostOnlyMatch.FindStringSubmatch(output)
	if len(matches) != 2 {
		panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) == 2:", createdHostOnlyMatch))
	}

	return matches[1], nil
}

func (n Networks) configureHostOnly(name, gateway, netmask string) error {
	args := []string{"hostonlyif", "ipconfig", name}

	if len(gateway) > 0 {
		args = append(args, []string{"--ip", gateway, "--netmask", netmask}...)
	} else {
		args = append(args, "--dhcp")
	}

	_, err := n.driver.ExecuteComplex(args, driver.ExecuteOpts{})

	return err
}

func (n Networks) cleanUpPartialHostOnlyCreate(name string) {
	_, err := n.driver.ExecuteComplex([]string{
		"hostonlyif",
		"remove",
		name,
	}, driver.ExecuteOpts{})
	if err != nil {
		n.logger.Error("vm.network.Networks",
			"Failed to clean up partially created host-only network '%s': %s", name, err)
	}
}
