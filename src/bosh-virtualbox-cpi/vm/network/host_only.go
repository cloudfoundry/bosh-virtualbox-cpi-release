package network

import (
	"bosh-virtualbox-cpi/driver"
	"fmt"
	"net"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
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
	return fmt.Sprintf("Host-only network '%s' (gw %s netmask %s)", n.name, n.ipAddress, n.networkMask)
}

func (n HostOnly) IsEnabled() bool {
	return n.status == "Up" || n.status == "Enabled"
}

func (n HostOnly) EnabledDescription() string {
	return fmt.Sprintf("have status '%s'", n.status)
}

func (n HostOnly) Enable() error {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	systemInfo, err := Networks{driver: n.driver, logger: logger}.NewSystemInfo()
	if err != nil {
		return err
	}

	var finalArgs []string
	if systemInfo.IsMacOSVbox7() {
		finalArgs = []string{"hostonlynet", "modify", fmt.Sprintf("--name=%s", n.name), "--enable"}
	} else {
		args := []string{"hostonlyif", "ipconfig", n.name}
		if len(n.ipAddress) > 0 {
			finalArgs = append(args, []string{"--ip", n.ipAddress, "--netmask", n.networkMask}...)
		} else {
			finalArgs = append(args, "--dhcp")
		}
	}

	_, err = n.driver.ExecuteComplex(finalArgs, driver.ExecuteOpts{})

	return err
}

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
