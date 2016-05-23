module BoshVirtualBoxCpi::Actions
  class RebootVm
    # @param [String] vm vm id that was once returned by {#create_vm}
    # @param [Optional, Hash] CPI specific options (e.g hard/soft reboot)
    def initialize(vm_manager, vm_id, cloud_props, logger)
      @vm_manager = vm_manager
      @vm_id = vm_id
      @cloud_props = cloud_props
      @logger = logger
    end

    def run
      vm = check_vm
      power_off(vm)
      power_on(vm)
    end

    private

    def check_vm
      @logger.info("Checking vm '#{@vm_id}'")
      @vm_manager.driver.vm_finder.find(@vm_id)
    end

    def power_off(vm)
      @logger.info("Powering off vm '#{vm.uuid}'")
      vm.halt if vm.running?
    end

    def power_on(vm)
      @logger.info("Powering on vm '#{vm.uuid}'")
      vm.start(@cloud_props.fetch("gui", false))
    end
  end
end
