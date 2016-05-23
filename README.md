# BOSH VirtualBox CPI

The BOSH VirtualBox CPI allows BOSH to manage *multiple* VirtualBox VMs/disks. It's very different from [BOSH Lite](https://github.com/cloudfoundry/bosh-lite).

- [Download VirtualBox 5+](https://www.virtualbox.org/wiki/Downloads)
- [Configuring Host-Only network](docs/networks-host-only.md) instructions to set up private VirtualBox network
- [Configuring NAT network](docs/networks-nat-network.md) instructions to set up public VirtualBox network

### Example

```
# Assumes vboxnet0 is configured as host-only network for 192.168.50.0/24
# and update bosh.yml with username and private key
$ bosh-init deploy manifests/bosh.yml

$ bosh target 192.168.50.6

# wget --content-disposition https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3232.4
$ bosh upload stemcell ~/Downloads/bosh-stemcell-3232.4-vsphere-esxi-ubuntu-trusty-go_agent.tgz

# wget --content-disposition https://bosh.io/d/github.com/cppforlife/zookeeper-release?v=0.0.1
$ bosh upload release ~/Downloads/zookeeper-release-0.0.1.tgz

$ bosh update cloud-config manifests/cloud.yml

# Replace director_uuid with the one from bosh status
$ bosh deployment manifests/zookeeper.yml

$ bosh deploy
```

### Problems

Pieces of `BoshVirtualBoxCpi::VirtualBox::Driver` class are based on Vagrant's VirtualBox driver.
