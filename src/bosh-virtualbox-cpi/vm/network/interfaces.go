package network

import (
	"net"
)

const (
	NATType        = "nat"
	NATNetworkType = "natnetwork"
	HostOnlyType   = "hostonly"
	BridgedType    = "bridged"
)

type Network interface {
	Name() string
	Description() string

	IsEnabled() bool
	EnabledDescription() string
	Enable() error

	IsDHCPEnabled() bool

	IPNet() *net.IPNet
}

var _ Network = HostOnly{}
var _ Network = NATNetwork{}
