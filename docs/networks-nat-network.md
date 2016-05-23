## Configuring network in 'NAT Network' mode (not 'NAT' mode) for internet access

NAT Network set up allows multiple VMs to be on the same network and
access network outside of the host (i.e. internet).
This is an experimental feature available in VirtualBox 4.3.0+
[documentation](http://www.virtualbox.org/manual/ch06.html#network_nat_service).

Set up a 'NAT Network' VirtualBox network:

1. Open VirtualBox
1. Choose VirtualBox > Preferences > Network
1. (Choose NAT Networks tab)
1. Create new network named `NatNetwork` (the default).
1. Check `Enable Network`
1. Uncheck `Supports DHCP`

### Example network configuration for BOSH manifest

```
networks:
- name: network-in-nat-network-mode
  subnets:
  - range:   10.0.13.0/24    # Network CIDR from configuration dialog
    gateway: 10.0.13.1       # has to end with .1
    dns:     ["8.8.8.8"]     # use public DNS
    static:
    - 10.0.2.4
    cloud_properties:
      name: mynatnet         # Network Name from configuration dialog
      type: natnetwork
```

Note: Tools using other than TCP or UDP protocols will not properly work. (e.g. `ping`)
