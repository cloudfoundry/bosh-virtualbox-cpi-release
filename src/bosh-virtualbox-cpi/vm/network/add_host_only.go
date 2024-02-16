package network

import (
	"bosh-virtualbox-cpi/driver"
	"fmt"
	"net"
	"regexp"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

var (
	createdHostOnlyMatch = regexp.MustCompile(`Interface '(.+)' was successfully created`)
)

func (n Networks) AddHostOnly(expectedName, gateway, netmask string) error {
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		return err
	}

	var createdName string
	if systemInfo.IsMacOSVboxV7OrLater() {
		if err := n.createHostOnlyNet(expectedName, gateway, netmask); err != nil {
			return err
		}
		createdName = expectedName
	} else {
		// Virtualox v6 or earlier choses itself the name of newly created
		// host-only networks
		createdName, err := n.createLegacyHostOnly()
		if err != nil {
			return err
		}
		if len(expectedName) > 0 && createdName != expectedName {
			n.cleanUpPartialHostOnlyCreate(createdName)
			return fmt.Errorf("expected created host-only network '%s' to have name '%s'. Have you made an incorrect guess?", createdName, expectedName)
		}
	}

	err = n.configureHostOnly(createdName, gateway, netmask)
	if err != nil {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return err
	}

	return nil
}

func (n Networks) createHostOnlyNet(name, gateway, netmask string) error {
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		return err
	}
	maskIP := net.ParseIP(netmask).To4()
	if maskIP == nil {
		return bosherr.Errorf("expected netmask to be valid IP v4 (got '%s')", netmask)
	}
	gwIP := net.ParseIP(gateway)
	if gwIP == nil {
		return bosherr.Errorf("expected gateway to be valid IP v4 (got '%s')", gateway)
	}

	mask := net.IPv4Mask(maskIP[0], maskIP[1], maskIP[2], maskIP[3])
	subnetFirstIP := &net.IPNet{
		IP:   gwIP,
		Mask: mask,
	}
	maskLength, _ := mask.Size()
	_, subnet, _ := net.ParseCIDR(fmt.Sprintf("%s/%v", gateway, maskLength))

	var lowerIP, upperIP net.IP
	if lowerIP, err = systemInfo.GetFirstIP(subnetFirstIP); err != nil {
		return err
	}
	if upperIP, err = systemInfo.GetLastIP(subnet); err != nil {
		return err
	}

	args := []string{"hostonlynet",
		"add", fmt.Sprintf("--name=%s", name),
		fmt.Sprintf("--netmask=%s", netmask), fmt.Sprintf("--lower-ip=%s", lowerIP.String()),
		fmt.Sprintf("--upper-ip=%s", upperIP.String()), "--disable"}

	// The output of the hostonlynet interface creation is empty. We need another solution to handle and verify the
	// VboxManage creation.
	_, err = n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
	if err != nil {
		return err
	}
	return nil
}

func (n Networks) createLegacyHostOnly() (string, error) {
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
	systemInfo, err := n.NewSystemInfo()
	if err != nil {
		return err
	}

	if !systemInfo.IsMacOSVboxV7OrLater() {
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

	args := []string{"hostonlyif", "remove"}
	if systemInfo.IsMacOSVboxV7OrLater() {
		args = append(args, fmt.Sprintf("--name=%s", name))
	} else {
		args = append(args, name)
	}

	_, err = n.driver.ExecuteComplex(args, driver.ExecuteOpts{})
	if err != nil {
		n.logger.Error("vm.network.Networks",
			"Failed to clean up partially created host-only network '%s': %s", name, err)
	}
}
