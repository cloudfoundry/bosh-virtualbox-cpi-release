package network

import (
	"fmt"
	"net"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type NATNetwork struct {
	driver driver.Driver

	name    string // e.g. NATNetwork
	enabled bool

	dhcpEnabled bool

	ipNet   *net.IPNet
	network string // e.g. 10.0.2.0/24
}

func (n NATNetwork) Name() string { return n.name }

func (n NATNetwork) Description() string {
	return fmt.Sprintf("NAT Network '%s' (network %s)", n.name, n.network)
}

func (n NATNetwork) IsEnabled() bool { return n.enabled }

func (n NATNetwork) EnabledDescription() string { return "be enabled" }

func (n NATNetwork) Enable() error {
	return fmt.Errorf("Enabling NAT Network is not implemented")
}

func (n NATNetwork) IsDHCPEnabled() bool { return n.dhcpEnabled }

func (n NATNetwork) IPNet() *net.IPNet { return n.ipNet }

func (n *NATNetwork) populateIPNet() error {
	_, ipNet, err := net.ParseCIDR(n.network)
	if err != nil {
		return fmt.Errorf("Unable to parse CIDR '%s' for network '%s': %s", n.network, n.name, err)
	}

	n.ipNet = ipNet

	return nil
}
