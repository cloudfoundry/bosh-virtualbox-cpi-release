module BoshVirtualBoxCpi::Virtualbox
  class CdromMounter
    def initialize(driver, vm, hot_plugger, logger)
      @driver = driver
      @vm = vm
      @hot_plugger = hot_plugger
      @logger = logger
    end

    def mount(iso_path)
      @logger.debug("virtualbox.cdrom_mounter.#{__method__} iso_path=#{iso_path}")

      @hot_plugger.hot_plug do
        @driver.execute(
          "storageattach", @vm.uuid,
          "--storagectl",  "IDE Controller",
          "--port",        "1",
          "--device",      "0",
          "--type",        "dvddrive",
          "--medium",      iso_path,
        )
      end
    end

    def unmount
      @logger.debug("virtualbox.cdrom_mounter.#{__method__}")

      @hot_plugger.hot_plug do
        @driver.execute(
          "storageattach", @vm.uuid,
          "--storagectl",  "IDE Controller",
          "--port",        "1",
          "--device",      "0",
          "--type",        "dvddrive",
          # 'emptydrive' removes medium from the drive
          # 'none' removes the device
          "--medium",      "emptydrive",
        )
      end
    end
  end
end
