module BoshVirtualBoxCpi::Virtualbox
  class NetworkConfigurer
    def initialize(driver, vm, logger)
      @driver = driver
      @vm = vm
      @logger = logger
    end

    def configure(network_options)
      @logger.debug("virtualbox.network_configurer.#{__method__} " +
        "network_names=#{network_options.map(&:inspect).inspect}")

      nic_to_no = build_nic_to_no(network_options)
      strategy  = configure_strategy

      nic_to_no.each do |nic, no|
        no ? strategy.set_nic(nic, no) : strategy.unset_nic(nic)
      end
    end

    def add_macs(network_options)
      @logger.debug("virtualbox.network_configurer.#{__method__} " +
        "network_names=#{network_options.map(&:inspect).inspect}")

      nic_to_no  = build_nic_to_no(network_options)
      nic_to_mac = read_nic_to_macs

      nic_to_no.each do |nic, no|
        next unless no
        raise "Missing mac address for '#{no.name}'" \
          unless no.mac = nic_to_mac[nic]
      end
    end

    private

    # https://0-forums.virtualbox.org.ilsprod.lib.neu.edu/viewtopic.php?f=1&t=53762
    def configure_strategy
      strategy = @vm.running? ? ForRunningVm : ForPoweredOffVm
      strategy.new(@driver, @vm, @logger)
    end

    # Attaching NICs to running VM is not allowed,
    # so 4 NICs will always be connected.
    MAX_NICS = 1

    def build_nic_to_no(network_options)
      raise ArgumentError, "Exceeded maximum # of NICs (#{MAX_NICS})" \
        if network_options.count > MAX_NICS
      Hash[(1..MAX_NICS).zip(network_options)]
    end

    def read_nic_to_macs
      @logger.debug("virtualbox.network_configurer.#{__method__}")
      macs = {}
      output = @driver.execute("showvminfo", @vm.uuid, "--machinereadable")
      output.split("\n").each do |line|
        if matcher = /^macaddress(\d+)="(.+?)"$/.match(line)
          macs[matcher[1].to_i] = matcher[2].to_s
        end
      end
      macs
    end

    # http://www.virtualbox.org/manual/ch06.html#network_nat_service
    # https://www.virtualbox.org/ticket/6176
    # `VBoxManage setextradata VM_NAME "VBoxInternal/Devices/pcnet/0/LUN#0/Config/Network" "172.23.24/24"`
    # `VBoxManage setextradata VM_NAME "VBoxInternal/Devices/pcnet/0/LUN#0/Config/DNSProxy" 1`
    class ForPoweredOffVm
      def initialize(*args)
        @driver, @vm, @logger = args
      end

      def set_nic(nic, network_option)
        name = network_option.cloud_name
        type = network_option.cloud_type

        @logger.debug("virtualbox.network_configurer.whenoff.#{__method__} " +
          "nic=#{nic} network_name=#{name} network_type=#{type}")

        opts = case type
          when "nat"        then ["nat"]
          when "natnetwork" then ["natnetwork", "--natnet#{nic}",          name]
          when "hostonly"   then ["hostonly",   "--hostonlyadapter#{nic}", name]
          else raise ArgumentError, "unknown '#{type}'"
        end

        @driver.execute("modifyvm", @vm.uuid, "--nic#{nic}", *opts)
      end

      def unset_nic(nic)
        @logger.debug("virtualbox.network_configurer.whenoff.#{__method__} nic=#{nic}")
        @driver.execute("modifyvm", @vm.uuid, "--nic#{nic}", "null")
      end
    end

    class ForRunningVm
      def initialize(*args)
        @driver, @vm, @logger = args
      end

      def set_nic(nic, network_option)
        name = network_option.cloud_name
        type = network_option.cloud_type

        @logger.debug("virtualbox.network_configurer.whenon.#{__method__} " +
          "nic=#{nic} network_name=#{name} network_type=#{type}")

        opts = case type
          when "nat"        then ["nat"]
          when "natnetwork" then ["natnetwork", name]
          when "hostonly"   then ["hostonly",   name]
          else raise ArgumentError, "unknown '#{type}'"
        end

        @driver.execute("controlvm", @vm.uuid, "nic#{nic}", *opts)
      end

      def unset_nic(nic)
        @logger.debug("virtualbox.network_configurer.whenon.#{__method__} nic=#{nic}")
        @driver.execute("controlvm", @vm.uuid, "nic#{nic}", "null")
      end
    end
  end
end
