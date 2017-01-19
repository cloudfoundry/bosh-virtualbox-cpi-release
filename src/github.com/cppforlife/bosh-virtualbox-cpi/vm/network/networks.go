package network

import (
	"fmt"
	"regexp"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	netKVMatch       = regexp.MustCompile(`^([a-zA-Z0-9]+):\s*(.+)?$`)
	netKVSpacedMatch = regexp.MustCompile(`^([a-zA-Z0-9\s]+):\s*(.+)?$`)
	hostOnlyMatch    = regexp.MustCompile(`Interface '(.+)' was successfully created`)
)

type Networks struct {
	driver driver.Driver
	logger boshlog.Logger
}

func NewNetworks(driver driver.Driver, logger boshlog.Logger) Networks {
	return Networks{driver, logger}
}

func (n Networks) AddNATNetwork(name string) error {
	output, err := n.driver.Execute(
		"natnetwork", "add",
		"--netname", name,
		"--network", "10.0.2.0/24",
		"--dhcp", "on",
	)
	if err != nil && !strings.Contains(output, "already exists") {
		return err
	}
	return nil
}

func (n Networks) AddHostOnly(name, gateway, netmask string) (bool, error) {
	// VB does not allow naming host-only networks, exit if it's not the first one
	if name != "vboxnet0" {
		return false, nil
	}

	output, err := n.driver.Execute("hostonlyif", "create")
	if err != nil {
		return true, err
	}

	matches := hostOnlyMatch.FindStringSubmatch(output)
	if len(matches) != 2 {
		panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) == 2:", hostOnlyMatch))
	}

	createdName := matches[1]

	if createdName != name {
		n.cleanUpPartialHostOnlyCreate(createdName)
		return true, fmt.Errorf("Expected created host-only network '%s' to have name '%s'", createdName, name)
	}

	args := []string{"hostonlyif", "ipconfig", name}

	if len(gateway) > 0 {
		args = append(args, []string{"--ip", gateway, "--netmask", netmask}...)
	} else {
		args = append(args, "--dhcp")
	}

	_, err = n.driver.Execute(args...)
	if err != nil {
		n.cleanUpPartialHostOnlyCreate(name)
		return true, err
	}

	return true, nil
}

func (n Networks) cleanUpPartialHostOnlyCreate(name string) {
	_, err := n.driver.Execute("hostonlyif", "remove", name)
	if err != nil {
		n.logger.Error("vm.network.Networks",
			"Failed to clean up partially created host-only network '%s': %s", name, err)
	}
}

func (n Networks) NATNetworks() ([]Network, error) {
	output, err := n.driver.Execute("list", "natnetworks")
	if err != nil {
		return nil, err
	}

	var nets []Network

	for _, netChunk := range n.outputChunks(output) {
		net := NATNetwork{driver: n.driver}

		for _, line := range strings.Split(netChunk, "\n") {
			if strings.Contains(line, "loopback mappings") {
				break
			}

			matches := netKVSpacedMatch.FindStringSubmatch(line)
			if len(matches) != 3 {
				panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) == 3: line '%s'", netKVSpacedMatch, line))
			}

			var err error

			switch matches[1] {
			// does not include all keys
			case "NetworkName":
				net.name = matches[2]
			case "DHCP Enabled":
				net.dhcpEnabled, err = n.toBool(matches[2])
			case "Network":
				net.network = matches[2]
			case "Enabled":
				net.enabled, err = n.toBool(matches[2])
			}

			if err != nil {
				return nil, err
			}
		}

		err := (&net).populateIPNet()
		if err != nil {
			return nil, err
		}

		nets = append(nets, net)
	}

	return nets, nil
}

func (n Networks) HostOnlys() ([]Network, error) {
	output, err := n.driver.Execute("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}

	var nets []Network

	for _, netChunk := range n.outputChunks(output) {
		net := HostOnly{driver: n.driver}

		for _, line := range strings.Split(netChunk, "\n") {
			matches := netKVMatch.FindStringSubmatch(line)
			if len(matches) != 3 {
				panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) == 3: line '%s'", netKVMatch, line))
			}

			var err error

			switch matches[1] {
			// does not include all keys
			case "Name":
				net.name = matches[2]
			case "DHCP":
				net.dhcp, err = n.toBool(matches[2])
			case "IPAddress":
				net.ipAddress = matches[2]
			case "NetworkMask":
				net.networkMask = matches[2]
			case "Status":
				net.status = matches[2]
			}

			if err != nil {
				return nil, err
			}
		}

		err := (&net).populateIPNet()
		if err != nil {
			return nil, err
		}

		nets = append(nets, net)
	}

	return nets, nil
}

func (n Networks) outputChunks(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n\n")
}

func (n Networks) toBool(s string) (bool, error) {
	switch s {
	case "Enabled", "Yes":
		return true, nil
	case "Disabled", "No":
		return false, nil
	default:
		return false, fmt.Errorf("Unknown boolean value '%s'", s)
	}
}
