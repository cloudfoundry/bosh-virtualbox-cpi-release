require "bosh_virtualbox_cpi/virtualbox/error"
require "bosh_virtualbox_cpi/virtualbox/vm"

module BoshVirtualBoxCpi::Virtualbox
  class VmFinder
    def initialize(driver, logger)
      @driver = driver
      @logger = logger
    end

    def find(uuid)
      @driver.execute("showvminfo", uuid)
      Vm.new(@driver, uuid, @logger)
    end
  end
end
