module BoshVirtualBoxCpi::Actions
  class DeleteVm
    # @param [String] vm vm id that was once returned by {#create_vm}
    def initialize(vm_manager, vm_id, logger)
      @vm_manager = vm_manager
      @vm_id = vm_id
      @logger = logger
    end

    def run
      vm = check_vm
      detach_persistent_disks(vm)
      delete_vm(vm)
    end

    private

    def check_vm
      @logger.info("Checking vm '#{@vm_id}'")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def detach_persistent_disks(vm)
      @logger.info("Detaching disks from vm '#{vm.uuid}'")
      artifact_keys = @vm_manager.list_artifacts(vm.uuid)
      artifact_keys.each do |key|
        detach_disk(vm, key) if key.end_with?("-disk-attachment.json")
      end
    end

    def detach_disk(vm, artifact_key)
      @logger.info("Detaching disk from vm '#{vm.uuid}'")

      contents = @vm_manager.get_artifact(vm.uuid, artifact_key)
      attachment = JSON.parse(contents)

      if attachment.fetch("type") == "persistent"
        disk_attacher = @vm_manager.driver.disk_attacher(vm)
        disk_attacher.detach(attachment.fetch("port_and_device"))
      end
    end

    def delete_vm(vm)
      @logger.info("Deleting vm '#{vm.uuid}'")
      vm.halt if vm.running?
      vm.delete
      @vm_manager.delete(vm.uuid)
    end
  end
end
