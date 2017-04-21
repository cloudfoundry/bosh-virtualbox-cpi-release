## Configuring network in 'Host-Only' mode for private networking

Set up a host-only VirtualBox network:

1. Open VirtualBox
1. Choose VirtualBox > Preferences > Network
1. Create new network named `vboxnet0`.
1. DHCP Server (tab) -> Uncheck `Enable Server`

Check that the `vboxnet0` network is configured on the host:

```
$ VBoxManage list hostonlyifs
$ ifconfig

_...other entries..._
vboxnet0: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> mtu 1500
	ether 0a:00:27:00:00:00
	inet 192.168.50.1 netmask 0xffffff00 broadcast 192.168.50.255

```

Finally, make sure you can ping the IP address `192.168.50.1`.
