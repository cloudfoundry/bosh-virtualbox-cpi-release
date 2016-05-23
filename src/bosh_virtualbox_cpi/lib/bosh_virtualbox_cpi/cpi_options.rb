require "yaml"

module BoshVirtualBoxCpi
  class CpiOptions
    attr_reader :host, :user, :private_key
    attr_reader :bin_path
    attr_reader :agent

    def self.from_hash(options, logger)
      credentials = options.values_at("host", "username", "private_key")

      bin_path  = options["bin_path"]
      store_dir = options["store_dir"]

      agent = options["agent"]

      new(*credentials, bin_path, store_dir, agent, logger)
    end

    def initialize(host, user, private_key, bin_path, store_dir, agent, logger)
      non_empty = lambda { |value, name|
        raise ArgumentError, "'#{name}' must not be empty" if value.to_s.empty?
        value
      }

      @host = non_empty.call(host, "host")
      @user = non_empty.call(user, "user")
      @private_key = non_empty.call(private_key, "private_key")

      @bin_path = non_empty.call(bin_path, "bin_path")
      @store_dir = non_empty.call(store_dir, "store_dir")

      @agent = agent
      @logger = logger
    end

    def local_access?
      @logger.info("cpi_options.#{__method__} remote_access=true")
      false
    end

    def stemcells_dir
      "#{@store_dir}/stemcells"
    end

    def vms_dir
      "#{@store_dir}/vms"
    end

    def disks_dir
      "#{@store_dir}/disks"
    end
  end
end
