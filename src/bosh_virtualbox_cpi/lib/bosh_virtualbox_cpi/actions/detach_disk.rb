module BoshVirtualBoxCpi::Actions
  class DetachDisk
    # @param [String] vm vm id that was once returned by {#create_vm}
    # @param [String] disk disk id that was once returned by {#create_disk}
    def initialize(vm_manager, disk_manager, vm_id, disk_id, type, logger)
      @vm_manager = vm_manager
      @disk_manager = disk_manager
      @vm_id = vm_id
      @disk_id = disk_id

      raise ArgumentError, "type is unknown" \
        unless %w(ephemeral persistent).include?(type)
      @type = type

      @logger = logger
    end

    def run
      vm = check_vm
      check_disk
      detach_disk(vm)
      agent_env = rebuild_agent_env(vm)
      mount_cdrom_with_agent_env(vm, agent_env)
    end

    private

    def check_vm
      @logger.info("Checking vm '#{@vm_id}'")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def check_disk
      @logger.info("Checking disk '#{@disk_id}'")
      raise "Could not find disk #{@disk_id}" \
        unless @disk_manager.exists?(@disk_id)
    end

    def detach_disk(vm)
      @logger.info("Detaching disk '#{@disk_id}' from vm '#{vm.uuid}'")

      contents = @vm_manager.get_artifact(vm.uuid, "#{@disk_id}-disk-attachment.json")
      attachment = JSON.parse(contents)

      disk_attacher = @vm_manager.driver.disk_attacher(vm)
      disk_attacher.detach(attachment.fetch("port_and_device"))
    end

    def rebuild_agent_env(vm)
      @logger.info("Rebuilding agent env for '#{vm.uuid}' without disk '#{@disk_id}'")
      contents = @vm_manager.get_artifact(vm.uuid, "env.json")
      BoshVirtualBoxCpi::AgentEnv.from_json(contents).tap do |env|
        env.send("remove_#{@type}_disk", @disk_id)
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
