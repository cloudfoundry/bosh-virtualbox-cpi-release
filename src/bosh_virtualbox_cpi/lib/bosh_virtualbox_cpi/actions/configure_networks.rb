require "bosh_virtualbox_cpi/network_option_collection"
require "bosh_virtualbox_cpi/agent_env"

module BoshVirtualBoxCpi::Actions
  class ConfigureNetworks
    # @param [String] vm vm id that was once returned by {#create_vm}
    # @param [Hash] networks list of networks and their settings needed for this VM,
    #               same as the networks argument in {#create_vm}
    def initialize(vm_manager, vm_id, networks, logger=Logger.new(STDERR))
      @vm_manager = vm_manager
      @vm_id = vm_id
      @network_options = \
        BoshVirtualBoxCpi::NetworkOptionCollection.from_hash(networks)
      @logger = logger
    end

    def run
      vm = check_vm
      configure_network(vm)
      agent_env = rebuild_agent_env(vm)
      mount_cdrom_with_agent_env(vm, agent_env)
    end

    private

    def check_vm
      @logger.info("Checking vm '#{@vm_id}'")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def configure_network(vm)
      @logger.info("Configuring network for '#{vm.uuid}'")
      configurer = @vm_manager.driver.network_configurer(vm)
      configurer.configure(@network_options)
    end

    def rebuild_agent_env(vm)
      @logger.info("Rebuilding agent env for '#{vm.uuid}'")

      configurer = @vm_manager.driver.network_configurer(vm)
      configurer.add_macs(@network_options)

      contents = @vm_manager.get_artifact(vm.uuid, "env.json")
      BoshVirtualBoxCpi::AgentEnv.from_json(contents).tap do |env|
        env.add_networks(@network_options)
      end
    end

    def mount_cdrom_with_agent_env(vm, agent_env)
      @logger.info("Mounting CDROM with updated agent env")

      @vm_manager.create_artifact(vm.uuid, "env.json", agent_env.as_json)
      @vm_manager.create_artifact(vm.uuid, "env.iso", agent_env.as_iso)

      cdrom = @vm_manager.driver.cdrom_mounter(vm)
      cdrom.mount(@vm_manager.artifact_path(vm.uuid, "env.iso"))
    end
  end
end
