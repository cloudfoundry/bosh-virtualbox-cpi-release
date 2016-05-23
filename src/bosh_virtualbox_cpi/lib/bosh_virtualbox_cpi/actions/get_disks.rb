module BoshVirtualBoxCpi::Actions
  class GetDisks
    def initialize(vm_manager, vm_id, logger)
      @vm_manager = vm_manager
      @vm_id = vm_id
      @logger = logger
    end

    def run
      vm = check_vm
      list_persistent_disks(vm)
    end

    private

    def check_vm
      @logger.info("Checking vm '#{@vm_id}'")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def list_persistent_disks(vm)
      @logger.info("Listing disks from vm '#{vm.uuid}'")

      artifact_keys = @vm_manager.list_artifacts(vm.uuid)

      artifact_keys.map { |key|
        if key.end_with?("-disk-attachment.json")
          contents = @vm_manager.get_artifact(vm.uuid, key)
          attachment = JSON.parse(contents)
          key if attachment.fetch("type") == "persistent"
        end
      }.compact
    end
  end
end
