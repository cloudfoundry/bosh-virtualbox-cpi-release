require "fileutils"

module BoshVirtualBoxCpi::Managers
  class Vm
    attr_reader :driver

    def initialize(vms_dir, runner, driver, logger)
      @vms_dir = vms_dir
      @runner = runner
      @driver = driver
      @logger = logger
    end

    def path(id)
      "#{@vms_dir}/#{id}"
    end

    def create(id)
      create_vms_dir
      create_vm_dir(id)
    end

    def delete(id)
      create_vms_dir
      @logger.debug("managers.vm.#{__method__} id=#{id}")
      @runner.execute!("rm", "-rf", path(id))
    end

    def artifact_path(id, key)
      "#{path(id)}/#{key}"
    end

    def create_artifact(id, key, contents)
      create_vms_dir
      create_vm_dir(id)
      @logger.debug("managers.vm.#{__method__} id=#{id} key=#{key}")
      @runner.put!(artifact_path(id, key), contents)
    end

    def get_artifact(id, key)
      @logger.debug("managers.vm.#{__method__} id=#{id} key=#{key}")
      @runner.get!(artifact_path(id, key))
    end

    def list_artifacts(id)
      create_vms_dir
      create_vm_dir(id)
      @logger.debug("managers.vm.#{__method__} id=#{id}")
      @runner.execute!("ls", "-1", path(id))[1].split("\n")
    end

    private

    def create_vms_dir
      @logger.debug("managers.vm.#{__method__} dir=#{@vms_dir}")
      @runner.execute!("mkdir", "-p", @vms_dir)
    end

    def create_vm_dir(id)
      @logger.debug("managers.vm.#{__method__} id=#{id}")
      @runner.execute!("mkdir", "-p", path(id))
    end
  end
end
