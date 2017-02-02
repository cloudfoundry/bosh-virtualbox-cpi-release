package vm

import (
	"fmt"
	gonet "net"

	bnet "github.com/cppforlife/bosh-virtualbox-cpi/vm/network"
)

type Host struct {
	networks bnet.Networks
}

func (h Host) EnableNetworks(nets Networks) error {
	for _, net := range nets {
		switch net.CloudPropertyType {
		case bnet.NATType:
			// do nothing

		case bnet.NATNetworkType:
			err := h.verifyOrCreateNATNetwork(net)
			if err != nil {
				return err
			}

		case bnet.HostOnlyType:
			err := h.verifyOrCreateHostOnly(net)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("Unknown network type: %s", net.CloudPropertyType)
		}
	}

	return nil
}

func (h Host) verifyOrCreateNATNetwork(net Network) error {
	return h.findAndVerifyNetwork(net, h.networks.NATNetworks, func() error {
		if len(net.IP) > 0 {
			return fmt.Errorf("Expected to find NAT Network '%s'", net.CloudPropertyName)
		}

		err := h.networks.AddNATNetwork(net.CloudPropertyName)
		if err != nil {
			return err
		}

		return h.findAndVerifyNetwork(net, h.networks.NATNetworks, nil)
	})
}

func (h Host) verifyOrCreateHostOnly(net Network) error {
	return h.findAndVerifyNetwork(net, h.networks.HostOnlys, func() error {
		canCreate, err := h.networks.AddHostOnly(net.CloudPropertyName, net.Gateway, net.Netmask)
		if err != nil {
			return err
		}

		if !canCreate {
			return fmt.Errorf("Expected to find Host-only network '%s'", net.CloudPropertyName)
		}

		return h.findAndVerifyNetwork(net, h.networks.HostOnlys, nil)
	})
}

func (h Host) findAndVerifyNetwork(net Network, listFunc func() ([]bnet.Network, error), notFoundFunc func() error) error {
	actualNets, err := listFunc()
	if err != nil {
		return err
	}

	for _, actualNet := range actualNets {
		if actualNet.Name() == net.CloudPropertyName {
			return h.verifyNetwork(net, actualNet)
		}
	}

	if notFoundFunc != nil {
		return notFoundFunc()
	}

	return fmt.Errorf("Expected to find network '%s'", net.CloudPropertyName)
}

func (h Host) verifyNetwork(net Network, actualNet bnet.Network) error {
	if !actualNet.IsEnabled() {
		return fmt.Errorf("Expected %s to %s", actualNet.Description(), actualNet.EnabledDescription())
	}

	if len(net.IP) == 0 {
		if !actualNet.IsDHCPEnabled() {
			return fmt.Errorf("Expected %s to have DHCP enabled", actualNet.Description())
		}
	} else {
		ip := gonet.ParseIP(net.IP)
		if ip == nil {
			return fmt.Errorf("Unable to parse IP address '%s' for network '%s'",
				net.IP, net.CloudPropertyName)
		}

		actualIPNet := actualNet.IPNet()

		if !actualIPNet.Contains(ip) {
			return fmt.Errorf("Expected IP address '%s' to fit within %s (%s)",
				net.IP, actualNet.Description(), actualIPNet.String())
		}

		actualNetmask := gonet.IP(actualIPNet.Mask).String()

		if actualNetmask != net.Netmask {
			return fmt.Errorf("Expected netmask '%s' to match %s netmask '%s'",
				net.IP, actualNet.Description(), actualNetmask)
		}

		// todo check gateway

		if actualNet.IsDHCPEnabled() {
			return fmt.Errorf("Expected %s to not have DHCP enabled", actualNet.Description())
		}
	}

	return nil
}
