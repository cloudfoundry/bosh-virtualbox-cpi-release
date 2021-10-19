## Configuring network in 'Bridged' mode for public networking

This will allow you to setup a BOSH director with the interface bridged over
to the local network. At the end of this, you should have a BOSH director
exposed to your LAN. Services deployed to the director should also be exposed
to the LAN.

**It is intended this is to be used for inhouse development or training purposes and is not considered production ready**

No Virtualbox changes should be needed. The following variables should be set
with the `internal_ip` set to a static IP on the LAN.

```yaml
internal_ip: 192.168.43.252
internal_gw: 192.168.43.3
internal_cidr: 192.168.43.0/24
outbound_network_name: NatNetwork
network_device: en0
```

The `network_device` name above can be found using `VBoxManage list bridgedifs`.
For example on macOS, a `network_device` of `"en0: Wi-Fi(Wireless)"` should
usually be necessary.

Add the following ops file to your `bosh create-env`:

```yaml
- type: replace
  path: /networks/name=default/subnets/0/cloud_properties?
  value:
    type: bridged
    name: ((network_device))
```

Adjust some routes. You'll need to let the BOSH Director VM know how to
route packets destinated for the container network.

You will also need to make sure any hosts on the LAN know how to route traffic
for services deployed to your Director via the BOSH Director's static IP. For
example:

```
sudo ip route add 10.244.0.0/16 via 192.168.43.252 dev en0
```
