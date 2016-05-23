require "bosh_virtualbox_cpi/runners/base"

module BoshVirtualBoxCpi::Runners
  class Timed < Base
    def initialize(runner, *args)
      super(*args)
      @runner = runner
    end

    %w(execute execute! upload! put! get!).each do |method|
      define_method(method) do |*args|
        log_time(method, args) { @runner.send(method, *args) }
      end
    end

    private

    def log_time(method, args, &blk)
      t1 = Time.now
      blk.call
    ensure
      t2 = Time.now

      args_to_log = args.first.to_s.end_with?(".iso") ? args.first : args

      # Careful logging all args since they
      # might contain iso contents, etc.
      logger.debug(
        "runners.timed.log_time " +
        "time=#{"%.03f" % (t2 - t1)}s " +
        "method=#{method} " +
        "args='#{args_to_log.inspect}'"
      )
    end
  end
end
