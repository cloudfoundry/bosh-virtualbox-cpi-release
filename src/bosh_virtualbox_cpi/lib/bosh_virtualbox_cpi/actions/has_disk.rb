module BoshVirtualBoxCpi::Actions
  class HasDisk
    def initialize(disk_manager, disk_id, logger)
      @disk_manager = disk_manager
      @disk_id = disk_id
      @logger = logger
    end

    def run
      @logger.info("Checking disk '#{@disk_id}'")
      @disk_manager.exists?(@disk_id)
    end
  end
end
