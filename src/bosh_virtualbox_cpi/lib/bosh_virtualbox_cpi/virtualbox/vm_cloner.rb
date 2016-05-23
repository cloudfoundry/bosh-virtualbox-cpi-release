require "securerandom"
require "bosh_virtualbox_cpi/virtualbox/error"
require "bosh_virtualbox_cpi/virtualbox/vm"

module BoshVirtualBoxCpi::Virtualbox
  class VmCloner
    PREPARED_SNAPSHOT_NAME = "prepared-clone"

    def initialize(driver, logger)
      @driver = driver
      @logger = logger
    end

    def prepare(vm)
      @logger.info("virtualbox.vm_cloner.#{__method__} vm=#{vm.uuid}")
      @driver.execute("snapshot", vm.uuid, "take", PREPARED_SNAPSHOT_NAME)
    end

    def clone(vm)
      @logger.info("virtualbox.vm_cloner.#{__method__} vm=#{vm.uuid}")

      cloned_vm_uuid = SecureRandom.uuid
      @driver.execute(
        "clonevm",    vm.uuid,
        "--snapshot", PREPARED_SNAPSHOT_NAME,
        "--options",  "link",
        "--name",     "vm-#{cloned_vm_uuid}", # extra non-conflicting
        "--uuid",     cloned_vm_uuid,
        "--register",
      )

      Vm.new(@driver, cloned_vm_uuid, @logger)
    end
  end
end
