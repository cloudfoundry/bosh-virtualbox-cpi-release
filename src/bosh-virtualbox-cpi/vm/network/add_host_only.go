package network

import (
	"bosh-virtualbox-cpi/driver"
	"fmt"
	"net"
	"regexp"
)

var (
	createdHostOnlyMatch    = regexp.MustCompile(`Interface '(.+)' was successfully created`)
	createdHostOnlyNetMatch = regexp.MustCompile(`Name:            vboxnet0`)
)

func (n Networks) AddHostOnly(name, gateway, netmask string) (bool, error) {
	// VB does not allow naming host-only networks inside version <= 6 , exit if it's not the first one
	if len(name) > 0 && name != "vboxnet0" {
		return false, nil
	}

	createdName, err := n.createHostOnly(gateway, netmask)

	if err != nil {
		return true, err
	}

	if len(name) > 0 && createdName != name {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return true, fmt.Errorf("expected created host-only network '%s' to have name '%s'", createdName, name)
	}

	err = n.configureHostOnly(createdName, gateway, netmask)
	if err != nil {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return true, err
	}

	return true, nil
}

func (n Networks) createHostOnly(gateway, netmask string) (string, error) {
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		return "", err
	}

	var matches []string
	var errorMessage string
	var expectedMatchesLen int

	if systemInfo.IsMacOSXVBoxSpecial6or7Case() {
		maskIP := net.ParseIP(netmask).To4()
		subnetFirstIP := &net.IPNet{
			IP:   net.ParseIP(gateway),
			Mask: net.IPv4Mask(maskIP[0], maskIP[1], maskIP[2], maskIP[3]),
		}
		maskLength, _ := net.IPv4Mask(maskIP[0], maskIP[1], maskIP[2], maskIP[3]).Size()
		_, subnet, _ := net.ParseCIDR(fmt.Sprintf("%s/%v", gateway, maskLength))

		lowerIp, err := systemInfo.GetFirstIP(subnetFirstIP)
		if err != nil {
			return "", err
		}
		upperIp, err := systemInfo.GetLastIP(subnet)
		if err != nil {
			return "", err
		}

		args := []string{"hostonlynet",
			"add", fmt.Sprintf("--name=%s", "vboxnet0"),
			fmt.Sprintf("--netmask=%s", netmask), fmt.Sprintf("--lower-ip=%s", lowerIp.String()),
			fmt.Sprintf("--upper-ip=%s", upperIp.String()), "--disable"}

		// The output of the hostonlynet interface creation is empty. We need another solution to handle and verify the
		// VboxManage creation.
		_, err = n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
		if err != nil {
			return "", err
		}

		args = []string{"list", "hostonlynets"}
		output, err := n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
		if err != nil {
			return "", err
		}

		matches = createdHostOnlyNetMatch.FindStringSubmatch(output)
		//Define the return value of the created Host only Adapter. We're only creating one adapter,
		//so we can also define the used name hard coded.
		if len(matches) == 1 {
			matches[0] = "vboxnet0"
		}

		errorMessage = fmt.Sprintf(
			"Internal inconsistency: Expected len(%s matches) == 1:",
			createdHostOnlyNetMatch,
		)
		expectedMatchesLen = 1
	} else {
		args := []string{"hostonlyif", "create"}
		output, err := n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
		if err != nil {
			return "", err
		}
		matches = createdHostOnlyMatch.FindStringSubmatch(output)
		errorMessage = fmt.Sprintf(
			"Internal inconsistency: Expected len(%s matches) == 2:",
			createdHostOnlyMatch,
		)
		expectedMatchesLen = 2
	}

	if len(matches) != expectedMatchesLen {
		panic(errorMessage)
	}

	return matches[expectedMatchesLen-1], nil
}

func (n Networks) configureHostOnly(name, gateway, netmask string) error {
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		return err
	}

	if systemInfo.IsMacOSXVBoxSpecial6or7Case() == false {
		args := []string{"hostonlyif", "ipconfig", name}

		if len(gateway) > 0 {
			args = append(args, []string{"--ip", gateway, "--netmask", netmask}...)
		} else {
			args = append(args, "--dhcp")
		}

		_, err := n.driver.ExecuteComplex(args, driver.ExecuteOpts{})

		return err
	} else {
		return nil
	}
}

func (n Networks) cleanUpPartialHostOnlyCreate(name string) {
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		n.logger.Error("vm.network.SystemInfo",
			"Failed to get the SystemInfo: %s", err)
	}

	args := []string{
		"hostonlyif",
		"remove",
		name,
	}
	if systemInfo.IsMacOSXVBoxSpecial6or7Case() {
		args = []string{
			"hostonlynet",
			"remove",
			fmt.Sprintf("--name=%s", name),
		}
	}

	_, err = n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
	if err != nil {
		n.logger.Error("vm.network.Networks",
			"Failed to clean up partially created host-only network '%s': %s", name, err)
	}
}
