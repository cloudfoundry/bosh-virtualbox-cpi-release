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
			err := newHostNetwork(net, natNetworksAdapter{h.networks}).Enable()
			if err != nil {
				return err
			}

		case bnet.HostOnlyType:
			err := newHostNetwork(net, hostOnlysAdapter{h.networks}).Enable()
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("Unknown network type: %s", net.CloudPropertyType)
		}
	}

	return nil
}

type hostNetwork struct {
	net     Network
	adapter netAdapter

	triedCreating bool
	triedEnabling bool
}

func newHostNetwork(net Network, adapter netAdapter) *hostNetwork {
	return &hostNetwork{net: net, adapter: adapter}
}

func (n *hostNetwork) Enable() error {
	actualNets, err := n.adapter.List()
	if err != nil {
		return err
	}

	for _, actualNet := range actualNets {
		if actualNet.Name() == n.net.CloudPropertyName {
			if !actualNet.IsEnabled() && !n.triedEnabling {
				_ = actualNet.Enable()
				n.triedEnabling = true
				return n.Enable()
			}
			return n.verify(actualNet)
		}
	}

	if !n.triedCreating {
		err := n.adapter.Create(n.net)
		if err != nil {
			return err
		}
		n.triedCreating = true
		return n.Enable()
	}

	return fmt.Errorf("Expected to find network '%s'", n.net.CloudPropertyName)
}

func (n hostNetwork) verify(actualNet bnet.Network) error {
	if !actualNet.IsEnabled() {
		return fmt.Errorf("Expected %s to %s",
			actualNet.Description(), actualNet.EnabledDescription())
	}

	if len(n.net.IP) == 0 {
		if !actualNet.IsDHCPEnabled() {
			return fmt.Errorf("Expected %s to have DHCP enabled", actualNet.Description())
		}
	} else {
		ip := gonet.ParseIP(n.net.IP)
		if ip == nil {
			return fmt.Errorf("Unable to parse IP address '%s' for network '%s'",
				n.net.IP, n.net.CloudPropertyName)
		}

		actualIPNet := actualNet.IPNet()

		if !actualIPNet.Contains(ip) {
			return fmt.Errorf("Expected IP address '%s' to fit within %s (%s)",
				n.net.IP, actualNet.Description(), actualIPNet.String())
		}

		actualNetmask := gonet.IP(actualIPNet.Mask).String()

		if actualNetmask != n.net.Netmask {
			return fmt.Errorf("Expected netmask '%s' to match %s netmask '%s'",
				n.net.IP, actualNet.Description(), actualNetmask)
		}

		// todo check gateway

		if actualNet.IsDHCPEnabled() {
			return fmt.Errorf("Expected %s to not have DHCP enabled", actualNet.Description())
		}
	}

	return nil
}

type netAdapter interface {
	List() ([]bnet.Network, error)
	Create(net Network) error
}

type natNetworksAdapter struct {
	bnet.Networks
}

func (n natNetworksAdapter) List() ([]bnet.Network, error) {
	return n.NATNetworks()
}

func (n natNetworksAdapter) Create(net Network) error {
	if len(net.IP) > 0 {
		return fmt.Errorf("Expected to find NAT Network '%s'", net.CloudPropertyName)
	}
	return n.AddNATNetwork(net.CloudPropertyName)
}

type hostOnlysAdapter struct {
	bnet.Networks
}

func (n hostOnlysAdapter) List() ([]bnet.Network, error) {
	return n.HostOnlys()
}

func (n hostOnlysAdapter) Create(net Network) error {
	canCreate, err := n.AddHostOnly(net.CloudPropertyName, net.Gateway, net.Netmask)
	if err != nil {
		return err
	} else if !canCreate {
		return fmt.Errorf("Expected to find Host-only network '%s'", net.CloudPropertyName)
	}
	return nil
}
