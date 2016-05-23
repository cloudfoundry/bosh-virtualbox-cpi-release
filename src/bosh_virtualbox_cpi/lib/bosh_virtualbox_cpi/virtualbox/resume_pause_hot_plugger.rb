module BoshVirtualBoxCpi::Virtualbox
  # VirtualBox as of 4.2.16 does not support real device hot plugging.
  # (http://www.youtube.com/watch?v=5c1m2BAg2Sc)
  class ResumePauseHotPlugger
    def initialize(driver, vm, logger)
      @driver = driver
      @vm = vm
      @logger = logger
    end

    def hot_plug(&blk)
      # http://dlc.sun.com.edgesuite.net/virtualbox/4.2.16/UserManual.pdf
      # Section 9.25: VirtualBox expert storage management
      @driver.execute(
        "setextradata", @vm.uuid,
        "VBoxInternal2/SilentReconfigureWhilePaused", "1",
      )

      if @vm.running?
        needs_to_resume = true
        @logger.info("virtualbox.resume_pause_hot_plugger.pause")
        @driver.execute("controlvm", @vm.uuid, "pause")
      end

      blk.call

      if needs_to_resume
        @logger.info("virtualbox.resume_pause_hot_plugger.resume")
        @driver.execute("controlvm", @vm.uuid, "resume")
      end
    end
  end
end
