module BoshVirtualBoxCpi::Actions
  class SetVmMetadata
    # @param [String] vm vm id that was once returned by {#create_vm}
    # @param [Hash] metadata metadata key/value pairs
    def initialize(vm_manager, vm_id, metadata, logger)
      @vm_manager = vm_manager
      @vm_id = vm_id
      @metadata = metadata
      @logger = logger
    end

    def run
      vm = check_vm
      set_metadata(vm)
    end

    private

    def check_vm
      @logger.info("Checking vm")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def set_metadata(vm)
      @logger.info("Setting metadata")
      @vm_manager.create_artifact(
        vm.uuid, "metadata.json", JSON.dump(@metadata))
    end
  end
end
