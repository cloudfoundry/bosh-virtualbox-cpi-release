## Configuring network in 'NAT' mode (not 'NAT Network' mode) for internet access

NAT set up allows individual VM to access network
outside of the host (i.e. internet). It does not allow VMs
under this type of network to communicate with each other.

No additional VirtualBox configuration is required.


### Example network configuration for BOSH manifest

Range (CIDR) for this type of network is picked in a very
specific way due to certain VirtualBox constraints:

- 1st network for VM: 10.0.2.0/24
- 2nd network for VM: 10.0.3.0/24
- ...

```
networks:
- name: network-in-nat-mode
  subnets:
  - range:   10.0.2.0/24     # has to be 10.0.x.0/24
    gateway: 10.0.2.2        # has to end with .2
    dns:     ["10.0.2.3"]    # has to end with .3
    static:
    - 10.0.2.4
    cloud_properties:
      type: nat
```


### Resources

> As more than one card of a virtual machine can be set up to use NAT,
> the first card is connected to the private network 10.0.2.0,
> the second card to the network 10.0.3.0 and so on...

[Blog post](http://geekynotebook.orangeonthewall.com/configure-static-ip-on-nat-in-oracle-virtualbox/)

> In NAT mode, the guest network interface is assigned to the IPv4
> range 10.0.x.0/24 by default where x corresponds to the instance
> of the NAT interface +2. So x is 2 when there is only one NAT
> instance active. In that case the guest is assigned to the
> address 10.0.2.15, the gateway is set to 10.0.2.2 and the name
> server can be found at 10.0.2.3.

[Forum post](https://forums.virtualbox.org/viewtopic.php?f=1&t=49066)
