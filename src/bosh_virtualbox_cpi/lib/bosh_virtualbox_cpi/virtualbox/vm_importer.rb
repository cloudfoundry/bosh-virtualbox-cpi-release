require "bosh_virtualbox_cpi/virtualbox/error"
require "bosh_virtualbox_cpi/virtualbox/retrier"
require "bosh_virtualbox_cpi/virtualbox/vm"

module BoshVirtualBoxCpi::Virtualbox
  class VmImporter
    def initialize(driver, retrier, logger)
      @driver = driver
      @retrier = retrier
      @logger = logger
    end

    def import(ovf_path)
      @retrier.times do
        output = @driver.execute("import", ovf_path)
        if output !~ /Suggested VM name "(.+?)"/
          raise BoshVirtualBoxCpi::Virtualbox::Error, "Couldn't find VM name in the output"
        end

        name = $1.to_s
        output = @driver.execute("list", "vms")
        if output =~ /^"#{Regexp.escape(name)}" \{(.+?)\}$/
          return Vm.new(@driver, $1.to_s, @logger)
        end

        raise BoshVirtualBoxCpi::Virtualbox::Error, "Failed to import #{ovf_path}"
      end
    end
  end
end
