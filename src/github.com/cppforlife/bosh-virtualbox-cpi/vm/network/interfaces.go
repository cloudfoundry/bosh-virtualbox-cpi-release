package network

import (
	"net"
)

const (
	NATType        = "nat"
	NATNetworkType = "natnetwork"
	HostOnlyType   = "hostonly"
)

type Network interface {
	Name() string
	Description() string

	IsEnabled() bool
	EnabledDescription() string

	IsDHCPEnabled() bool

	IPNet() *net.IPNet
}

var _ Network = HostOnly{}
var _ Network = NATNetwork{}
