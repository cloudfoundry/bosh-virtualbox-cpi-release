package network

import (
	"fmt"
	"net"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type HostOnly struct {
	driver driver.Driver

	name   string // e.g. vboxnet0
	status string // e.g. Up

	dhcp bool

	ipNet       *net.IPNet
	ipAddress   string // e.g. 192.168.56.1
	networkMask string // e.g. 255.255.255.0
}

func (n HostOnly) Name() string { return n.name }

func (n HostOnly) Description() string {
	return fmt.Sprintf("Host-only network '%s'", n.name)
}

func (n HostOnly) IsEnabled() bool { return n.status == "Up" }

func (n HostOnly) EnabledDescription() string { return "have status 'Up'" }

func (n HostOnly) IsDHCPEnabled() bool { return n.dhcp }

func (n HostOnly) IPNet() *net.IPNet { return n.ipNet }

func (n *HostOnly) populateIPNet() error {
	ip := net.ParseIP(n.ipAddress)
	if ip == nil {
		return fmt.Errorf("Unable to parse IP address '%s' for network '%s'", n.ipAddress, n.name)
	}

	maskIP := net.ParseIP(n.networkMask)
	if maskIP == nil {
		return fmt.Errorf("Unable to parse network mask '%s' for network '%s'", n.networkMask, n.name)
	}

	n.ipNet = &net.IPNet{IP: ip, Mask: net.IPMask(maskIP)}

	return nil
}
