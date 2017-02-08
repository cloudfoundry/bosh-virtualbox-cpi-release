package action

import (
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type Networks map[string]Network

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`

	DNS     []string `json:"dns"`
	Default []string `json:"default"`

	CloudProperties NetworkCloudProperties `json:"cloud_properties"`
}

type NetworkCloudProperties struct {
	Name string
	Type string
}

func (ns Networks) AsVMNetworks() bvm.Networks {
	nets := bvm.Networks{}

	for netName, network := range ns {
		net := bvm.Network{
			Type: network.Type,

			IP:      network.IP,
			Netmask: network.Netmask,
			Gateway: network.Gateway,

			DNS:     network.DNS,
			Default: network.Default,

			CloudPropertyName: network.CloudProperties.Name,
			CloudPropertyType: network.CloudProperties.Type,
		}

		if net.CloudPropertyType == "" {
			net.CloudPropertyType = "hostonly"
		}

		nets[netName] = net
	}

	return nets
}
