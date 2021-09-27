package vm

import (
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"
)

type Networks map[string]Network

type Network struct {
	net   apiv1.Network
	props NetworkCloudProps
}

type NetworkCloudProps struct {
	Name string
	Type string
}

func NewNetworks(nets apiv1.Networks) (Networks, error) {
	newNets := Networks{}

	for name, net := range nets {
		newNet, err := NewNetwork(net)
		if err != nil {
			return newNets, err
		}
		newNets[name] = newNet
	}

	return newNets, nil
}

func NewNetwork(net apiv1.Network) (Network, error) {
	var props NetworkCloudProps

	err := net.CloudProps().As(&props)
	if err != nil {
		return Network{}, err
	}

	if props.Type == "" {
		props.Type = "hostonly"
	}

	return Network{net, props}, nil
}

func (n Network) IP() string      { return n.net.IP() }
func (n Network) Netmask() string { return n.net.Netmask() }
func (n Network) Gateway() string { return n.net.Gateway() }

func (n Network) SetMAC(mac string) { n.net.SetMAC(mac) }

func (n Network) CloudPropertyName() string { return n.props.Name }
func (n Network) CloudPropertyType() string { return n.props.Type }

func (ns Networks) AsNetworks() apiv1.Networks {
	newNets := apiv1.Networks{}

	for name, net := range ns {
		newNets[name] = net.net
	}

	return newNets
}
