require "bosh_virtualbox_cpi/virtualbox/error"

module BoshVirtualBoxCpi::Virtualbox
  class Retrier
    def initialize(logger)
      @logger = logger
    end

    def times(times=10, sleep=5, &blk)
      blk.call
    rescue BoshVirtualBoxCpi::Virtualbox::Error => e
      times -= 1
      if times.zero?
        @logger.info("virtualbox.vm_importer.retry.failed error=#{e.inspect}")
        raise
      else
        @logger.info("virtualbox.vm_importer.retry try=#{times} error=#{e.inspect}")
        sleep(sleep)
        retry
      end
    end
  end
end
