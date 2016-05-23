module BoshVirtualBoxCpi::Actions
  class DeleteStemcell
    # @param [String] stemcell stemcell id that was once returned by {#create_stemcell}
    def initialize(stemcell_manager, stemcell_id, logger)
      @stemcell_manager = stemcell_manager
      @stemcell_id = stemcell_id
      @logger = logger
    end

    def run
      check_stemcell
      delete_vm
      delete_stemcell
    end

    private

    def check_stemcell
      @logger.info("Checking stemcell '#{@stemcell_id}'")
      raise "Could not find stemcell #{@stemcell_id}" \
        unless @stemcell_manager.exists?(@stemcell_id)
    end

    def delete_vm
      @logger.info("Deleting stemcell VM '#{@stemcell_id}'")
      vm_id = @stemcell_manager.get_artifact(@stemcell_id, "vm-id")
      vm    = @stemcell_manager.driver.vm_finder.find(vm_id)
      @stemcell_manager.driver.vm_cloner.clone(vm)
    end

    def delete_stemcell
      @logger.info("Deleting stemcell '#{@stemcell_id}'")
      @stemcell_manager.delete(@stemcell_id)
    end
  end
end
