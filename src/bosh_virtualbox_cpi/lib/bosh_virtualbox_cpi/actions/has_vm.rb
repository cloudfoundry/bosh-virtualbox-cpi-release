module BoshVirtualBoxCpi::Actions
  class HasVm
    # @param [String] vm vm id that was once returned by {#create_vm}
    def initialize(vm_manager, vm_id, logger)
      @vm_manager = vm_manager
      @vm_id = vm_id
      @logger = logger
    end

    # @return [Boolean] True if the vm exists
    def run
      @logger.info("Checking vm")
      !!@vm_manager.driver.vm_finder.find(@vm_id)
    rescue BoshVirtualBoxCpi::Virtualbox::Error
      false
    end
  end
end
