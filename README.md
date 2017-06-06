# BOSH VirtualBox CPI

The BOSH VirtualBox CPI allows BOSH to manage *multiple* VirtualBox VMs/disks.

- [Download VirtualBox 5+](https://www.virtualbox.org/wiki/Downloads)
- [Cloud properties](docs/cloud-props.md)
- [Configuring Host-Only network](docs/networks-host-only.md) instructions to set up private VirtualBox network
- [Configuring NAT network](docs/networks-nat-network.md) instructions to set up public VirtualBox network

See [bosh-deployment's BOSH Lite on VirtualBox](https://github.com/cloudfoundry/bosh-deployment/blob/master/docs/bosh-lite-on-vbox.md) or [Concourse deployment](https://github.com/cppforlife/concourse-deployment) for example usage.

## TODO

- Aggressive VM deletion

```
CPI 'delete_vm' method responded with error: CmdError{"type":"Bosh::Clouds::CloudError","message":"Deleting vm 'vm-8b33e9d9-525f-49a9-6e1e-b156194ca0fe': Determining controller name: Retried '30' times: Running command: 'VBoxManage showvminfo vm-8b33e9d9-525f-49a9-6e1e-b156194ca0fe --machinereadable', stdout: '', stderr: 'VBoxManage: error: Could not find a registered machine named 'vm-8b33e9d9-525f-49a9-6e1e-b156194ca0fe'\nVBoxManage: error: Details: code VBOX_E_OBJECT_NOT_FOUND (0x80bb0001), component VirtualBoxWrap, interface IVirtualBox, callee nsISupports\nVBoxManage: error: Context: \"FindMachine(Bstr(VMNameOrUuid).raw(), machine.asOutParam())\" at line 2781 of file VBoxManageInfo.cpp\n': exit status 1","ok_to_retry":false}
```
