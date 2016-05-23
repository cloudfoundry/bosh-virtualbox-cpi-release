require "securerandom"

module BoshVirtualBoxCpi::Managers
  class Disk
    PREFIX = "disk"

    attr_reader :driver

    def initialize(disks_dir, runner, driver, logger)
      @disks_dir = disks_dir
      @runner = runner
      @driver = driver
      @logger = logger
    end

    def path(id)
      "#{@disks_dir}/#{id}"
    end

    def create
      create_disks_dir
      id = "#{PREFIX}-#{SecureRandom.uuid}"
      @logger.debug("managers.disk.#{__method__} id=#{id}")
      @runner.execute!("mkdir", "-p", path(id))
      id
    end

    def exists?(id)
      @logger.debug("managers.disk.#{__method__} id=#{id}")
      exit_status, _ = @runner.execute("ls", path(id))
      exit_status.zero?
    end

    def delete(id)
      create_disks_dir
      @logger.debug("managers.disk.#{__method__} id=#{id}")
      @runner.execute!("rm", "-rf", path(id))
    end

    private

    def create_disks_dir
      @logger.debug("managers.disk.#{__method__} dir=#{@disks_dir}")
      @runner.execute!("mkdir", "-p", @disks_dir)
    end
  end
end
