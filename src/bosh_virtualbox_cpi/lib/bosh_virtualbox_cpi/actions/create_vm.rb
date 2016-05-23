require "bosh_virtualbox_cpi/network_option_collection"
require "bosh_virtualbox_cpi/agent_env"

module BoshVirtualBoxCpi::Actions
  class CreateVm
    # Sample networking config:
    #  {"network_a" =>
    #    {
    #      "netmask"          => "255.255.248.0",
    #      "ip"               => "172.30.41.40",
    #      "gateway"          => "172.30.40.1",
    #      "dns"              => ["172.30.22.153", "172.30.22.154"],
    #      "cloud_properties" => {"name" => "VLAN444"}
    #    }
    #  }
    def initialize(stemcell_manager, vm_manager, agent_options,
                   agent_id, stemcell_id, cloud_props,
                   networks, disk_locality=nil, env=nil, logger)
      @stemcell_manager = stemcell_manager
      @vm_manager = vm_manager
      @agent_options = agent_options
      @agent_id = agent_id
      @stemcell_id = stemcell_id
      @cloud_props = cloud_props
      @network_options = \
        BoshVirtualBoxCpi::NetworkOptionCollection.from_hash(networks)
      @disk_locality = disk_locality
      @env = env
      @logger = logger
    end

    def run
      check_stemcell
      vm = clone_vm
      set_name(vm)
      configure_network(vm)
      agent_env = build_agent_env(vm)
      mount_cdrom_with_agent_env(vm, agent_env)
      vm.uuid
    rescue Exception => e
      clean_up_partial_vm(vm)
      raise
    end

    private

    def check_stemcell
      @logger.info("Checking stemcell '#{@stemcell_id}'")
      raise "Could not find stemcell #{@stemcell_id}" \
        unless @stemcell_manager.exists?(@stemcell_id)
    end

    def clone_vm
      @logger.info("Cloning VM from stemcell VM")
      stemcell_vm_id = @stemcell_manager.get_artifact(@stemcell_id, "vm-id")
      stemcell_vm    = @stemcell_manager.driver.vm_finder.find(stemcell_vm_id)
      @vm_manager.driver.vm_cloner.clone(stemcell_vm)
    end

    def set_name(vm)
      @logger.info("Setting name for '#{vm.uuid}'")
      vm.name = "vm-#{vm.uuid}"
    end

    def configure_network(vm)
      @logger.info("Configuring network for '#{vm.uuid}'")
      configurer = @vm_manager.driver.network_configurer(vm)
      configurer.configure(@network_options)
    end

    def build_agent_env(vm)
      @logger.info("Building agent env for '#{vm.uuid}'")

      configurer = @vm_manager.driver.network_configurer(vm)
      configurer.add_macs(@network_options)

      BoshVirtualBoxCpi::AgentEnv.new.tap do |env|
        env.vm_id = vm.uuid
        env.name = vm.uuid
        env.env = @env
        env.agent_id = @agent_id
        env.agent_env = @agent_options
        env.add_networks(@network_options)
        env.add_empty_disks
      end
    end

    def mount_cdrom_with_agent_env(vm, agent_env)
      @logger.info("Mounting CDROM with agent env")

      @vm_manager.create_artifact(vm.uuid, "env.json", agent_env.as_json)
      @vm_manager.create_artifact(vm.uuid, "env.iso", agent_env.as_iso)

      cdrom = @vm_manager.driver.cdrom_mounter(vm)
      cdrom.mount(@vm_manager.artifact_path(vm.uuid, "env.iso"))
    end

    def clean_up_partial_vm(vm)
      return unless vm
      vm.delete
      @vm_manager.delete(vm.uuid) if vm.uuid
    end
  end
end
