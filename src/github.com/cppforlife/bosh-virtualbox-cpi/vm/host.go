package vm

import (
	"fmt"
	gonet "net"

	bnet "github.com/cppforlife/bosh-virtualbox-cpi/vm/network"
)

type Host struct {
	networks bnet.Networks
}

func (h Host) FindNetwork(net Network) (bnet.Network, error) {
	switch net.CloudPropertyType() {
	case bnet.NATType:
		return nil, fmt.Errorf("NAT networks cannot be searched")

	case bnet.NATNetworkType:
		return newHostNetwork(net, natNetworksAdapter{h.networks}).Find()

	case bnet.HostOnlyType:
		return newHostNetwork(net, hostOnlysAdapter{h.networks}).Find()

	case bnet.BridgedType:
		return newHostNetwork(net, bridgedNetworksAdapter{h.networks}).Find()

	default:
		return nil, fmt.Errorf("Unknown network type: %s", net.CloudPropertyType())
	}
}

func (h Host) EnableNetworks(nets Networks) error {
	for _, net := range nets {
		switch net.CloudPropertyType() {
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
			return fmt.Errorf("Unknown network type: %s", net.CloudPropertyType())
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

func (n *hostNetwork) Find() (bnet.Network, error) {
	actualNets, err := n.adapter.List()
	if err != nil {
		return nil, err
	}

	for _, actualNet := range actualNets {
		if n.adapter.Matches(n.net, actualNet) {
			return actualNet, nil
		}
	}

	return nil, fmt.Errorf("Expected to find network '%s'", n.net.CloudPropertyName())
}

func (n *hostNetwork) Enable() error {
	actualNets, err := n.adapter.List()
	if err != nil {
		return err
	}

	for _, actualNet := range actualNets {
		if n.adapter.Matches(n.net, actualNet) {
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

	return fmt.Errorf("Expected to find network '%s'", n.net.CloudPropertyName())
}

func (n hostNetwork) verify(actualNet bnet.Network) error {
	if !actualNet.IsEnabled() {
		return fmt.Errorf("Expected %s to %s",
			actualNet.Description(), actualNet.EnabledDescription())
	}

	if len(n.net.IP()) == 0 {
		if !actualNet.IsDHCPEnabled() {
			return fmt.Errorf("Expected %s to have DHCP enabled", actualNet.Description())
		}
	} else {
		ip := gonet.ParseIP(n.net.IP())
		if ip == nil {
			return fmt.Errorf("Unable to parse IP address '%s' for network '%s'",
				n.net.IP(), n.net.CloudPropertyName())
		}

		actualIPNet := actualNet.IPNet()

		if !actualIPNet.Contains(ip) {
			return fmt.Errorf("Expected IP address '%s' to fit within %s", n.net.IP(), actualNet.Description())
		}

		actualNetmask := gonet.IP(actualIPNet.Mask).String()

		if actualNetmask != n.net.Netmask() {
			return fmt.Errorf("Expected netmask '%s' to match %s netmask '%s'",
				n.net.IP(), actualNet.Description(), actualNetmask)
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
	Create(Network) error
	Matches(Network, bnet.Network) bool
}

type natNetworksAdapter struct {
	bnet.Networks
}

func (n natNetworksAdapter) List() ([]bnet.Network, error) {
	return n.NATNetworks()
}

func (n natNetworksAdapter) Create(net Network) error {
	if len(net.IP()) > 0 {
		return fmt.Errorf("Expected to find NAT Network '%s'", net.CloudPropertyName())
	}
	return n.AddNATNetwork(net.CloudPropertyName())
}

func (n natNetworksAdapter) Matches(net Network, actualNet bnet.Network) bool {
	return net.CloudPropertyName() == actualNet.Name()
}

type hostOnlysAdapter struct {
	bnet.Networks
}

func (n hostOnlysAdapter) List() ([]bnet.Network, error) {
	return n.HostOnlys()
}

func (n hostOnlysAdapter) Create(net Network) error {
	canCreate, err := n.AddHostOnly(net.CloudPropertyName(), net.Gateway(), net.Netmask())
	if err != nil {
		return err
	} else if !canCreate {
		return fmt.Errorf("Expected to find Host-only network '%s'", net.CloudPropertyName())
	}
	return nil
}

func (n hostOnlysAdapter) Matches(net Network, actualNet bnet.Network) bool {
	if len(net.CloudPropertyName()) > 0 {
		return net.CloudPropertyName() == actualNet.Name()
	}

	actualIP := gonet.IP(actualNet.IPNet().IP).String()
	actualNetmask := gonet.IP(actualNet.IPNet().Mask).String()

	return actualNetmask == net.Netmask() && actualIP == net.Gateway()
}

type bridgedNetworksAdapter struct {
	bnet.Networks
}

func (n bridgedNetworksAdapter) List() ([]bnet.Network, error) {
	return n.BridgedNetworks()
}

func (n bridgedNetworksAdapter) Create(net Network) error {
	return fmt.Errorf("Expected to find bridged network '%s'", net.CloudPropertyName())
}

func (n bridgedNetworksAdapter) Matches(net Network, actualNet bnet.Network) bool {
	if len(net.CloudPropertyName()) > 0 {
		return net.CloudPropertyName() == actualNet.Name()
	}

	actualIP := gonet.IP(actualNet.IPNet().IP).String()
	actualNetmask := gonet.IP(actualNet.IPNet().Mask).String()

	return actualNetmask == net.Netmask() && actualIP == net.Gateway()
}
