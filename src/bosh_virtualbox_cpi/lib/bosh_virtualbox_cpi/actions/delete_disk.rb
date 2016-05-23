module BoshVirtualBoxCpi::Actions
  class DeleteDisk
    # Deletes a disk
    # Will raise an exception if the disk is attached to a VM
    # @param [String] disk disk id that was once returned by {#create_disk}
    def initialize(disk_manager, disk_id, logger)
      @disk_manager = disk_manager
      @disk_id = disk_id
      @logger = logger
    end

    def run
      check_disk
      delete_disk
    end

    private

    def check_disk
      @logger.info("Checking disk '#{@disk_id}'")
      raise "Could not find disk #{@disk_id}" unless @disk_manager.exists?(@disk_id)
    end

    def delete_disk
      @logger.info("Deleting disk '#{@disk_id}'")
      @disk_manager.delete(@disk_id)
    end
  end
end
