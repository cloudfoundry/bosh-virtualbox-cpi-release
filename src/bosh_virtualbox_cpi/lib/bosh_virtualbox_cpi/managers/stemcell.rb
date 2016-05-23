require "securerandom"

module BoshVirtualBoxCpi::Managers
  class Stemcell
    PREFIX = "sc"

    attr_reader :driver

    def initialize(stemcells_dir, runner, driver, logger)
      @stemcells_dir = stemcells_dir
      @driver = driver
      @runner = runner
      @logger = logger
    end

    def path(id)
      "#{@stemcells_dir}/#{id}"
    end

    def create(dir)
      create_stemcells_dir
      id = "#{PREFIX}-#{SecureRandom.uuid}"
      @logger.debug("managers.stemcell.#{__method__} id=#{id} dir=#{dir}")
      @runner.upload!(dir, path(id))
      id
    end

    def exists?(id)
      @logger.debug("managers.stemcell.#{__method__} id=#{id}")
      exit_status, _ = @runner.execute("ls", path(id))
      exit_status.zero?
    end

    def delete(id)
      create_stemcells_dir
      @logger.debug("managers.stemcell.#{__method__} id=#{id}")
      @runner.execute!("rm", "-rf", path(id))
    end

    def artifact_path(id, key)
      "#{path(id)}/#{key}"
    end

    def create_artifact(id, key, contents)
      create_stemcells_dir
      create_stemcell_dir(id)
      @logger.debug("managers.stemcell.#{__method__} id=#{id} key=#{key}")
      @runner.put!(artifact_path(id, key), contents)
    end

    def get_artifact(id, key)
      @logger.debug("managers.stemcell.#{__method__} id=#{id} key=#{key}")
      @runner.get!(artifact_path(id, key))
    end

    private

    def create_stemcells_dir
      @logger.debug("managers.stemcell.#{__method__} dir=#{@stemcells_dir}")
      @runner.execute!("mkdir", "-p", @stemcells_dir)
    end

    def create_stemcell_dir(id)
      @logger.debug("managers.stemcell.#{__method__} id=#{id}")
      @runner.execute!("mkdir", "-p", path(id))
    end
  end
end
