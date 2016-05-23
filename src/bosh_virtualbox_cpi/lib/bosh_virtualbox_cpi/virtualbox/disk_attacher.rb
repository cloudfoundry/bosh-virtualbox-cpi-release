require "bosh_virtualbox_cpi/virtualbox/error"

module BoshVirtualBoxCpi::Virtualbox
  class DiskAttacher
    def initialize(driver, vm, hot_plugger, logger)
      @driver = driver
      @vm = vm
      @hot_plugger = hot_plugger
      @logger = logger
    end

    def attach(path)
      @logger.debug("virtualbox.disk_attacher.#{__method__} uuid=#{@vm.uuid} path=#{path}")

      raise BoshVirtualBoxCpi::Virtualbox::Error, \
        "Failed to obtain port and device to SCSI Controller" \
          if (pads = read_empty_scsi_ports_and_devices).empty?

      port_and_device = pads.first
      @logger.debug("virtualbox.disk_attacher.#{__method__} " +
        "uuid=#{@vm.uuid} path=#{path} port_and_device=#{port_and_device.inspect}")

      @hot_plugger.hot_plug do
        @driver.execute(
          "storageattach", @vm.uuid,
          "--storagectl",  "SCSI Controller",
          "--port",        port_and_device.first.to_s,
          "--device",      port_and_device.last.to_s,
          "--type",        "hdd",
          "--medium",      "#{path}/disk.vmdk",
          "--mtype",       "normal",
        )
      end

      port_and_device
    end

    def detach(port_and_device)
      @logger.debug("virtualbox.disk_attacher.#{__method__} " +
        "uuid=#{@vm.uuid} port_and_device=#{port_and_device}")

      @hot_plugger.hot_plug do
        @driver.execute(
          "storageattach", @vm.uuid,
          "--storagectl",  "SCSI Controller",
          "--port",        port_and_device.first.to_s,
          "--device",      port_and_device.last.to_s,
          "--type",        "hdd",
          "--medium",      "none", # removes
        )
      end
    end

    private

    def read_empty_scsi_ports_and_devices
      @logger.debug("virtualbox.disk_attacher.#{__method__} uuid=#{@vm.uuid}")

      sleep(5)
      output = @driver.execute("showvminfo", @vm.uuid, "--machinereadable")

      ports_and_devices = []
      output.split("\n").each do |line|
        if line =~ /^"SCSI Controller-(\d+)-(\d+)"="none"$/
          ports_and_devices << [$1, $2]
        end
      end

      @logger.debug("virtualbox.disk_attacher.#{__method__} " +
        "uuid=#{@vm.uuid} ports_and_devices=#{ports_and_devices.inspect}")
      ports_and_devices
    end
  end
end
