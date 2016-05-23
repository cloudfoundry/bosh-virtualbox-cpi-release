require "bosh_virtualbox_cpi/cpi_options"
require "bosh_virtualbox_cpi/resolved_dir"
require "bosh_virtualbox_cpi/runners"
require "bosh_virtualbox_cpi/virtualbox"
require "bosh_virtualbox_cpi/managers"
require "bosh_virtualbox_cpi/actions"

module BoshVirtualBoxCpi
  class Cpi
    def self.default_logger
      Logger.new(STDERR)
    end

    def initialize(options, logger=nil)
      @logger  = logger || self.class.default_logger
      @options = CpiOptions.from_hash(options, @logger)

      if @options.local_access?
        runner = Runners::Local.new(@logger)
      else
        runner = Runners::Remote.new(
          @options.host,
          @options.user,
          @options.private_key,
          @logger,
        )
      end

      runner = Runners::Timed.new(runner, @logger)
      retrier = Virtualbox::Retrier.new(@logger)
      driver = Virtualbox::Driver.new(runner, @options.bin_path, retrier, @logger)

      stemcells_dir = ResolvedDir.new(@options.stemcells_dir, runner)
      @stemcell_manager = Managers::Stemcell.new(stemcells_dir, runner, driver, @logger)

      vms_dir = ResolvedDir.new(@options.vms_dir, runner)
      @vm_manager = Managers::Vm.new(vms_dir, runner, driver, @logger)

      disks_dir = ResolvedDir.new(@options.disks_dir, runner)
      @disk_manager = Managers::Disk.new(disks_dir, runner, driver, @logger)
    end

    def create_stemcell(*args)
      Actions::CreateStemcell.new(@stemcell_manager, *args, @logger).run
    end

    def delete_stemcell(*args)
      Actions::DeleteStemcell.new(@stemcell_manager, *args, @logger).run
    end

    def create_vm(agent_id, stemcell_id, cloud_props, networks, disk_locality=nil, env=nil)
      vm_id = Actions::CreateVm.new(
        @stemcell_manager, @vm_manager, @options.agent,
        agent_id, stemcell_id, cloud_props,
        networks, disk_locality, env, @logger,
      ).run

      disk_id = Actions::CreateDisk.new(@disk_manager, cloud_props["disk"], {}, nil, @logger).run

      # create_vm CPI action *must* attach ephemeral disk
      # *before* powering on VM, otherwise, bosh_agent
      # will not properly bootstrap environment.
      Actions::AttachDisk.new(@vm_manager, @disk_manager, vm_id, disk_id, "ephemeral", @logger).run

      Actions::RebootVm.new(@vm_manager, vm_id, cloud_props, @logger).run

      vm_id
    end

    def delete_vm(*args)
      Actions::DeleteVm.new(@vm_manager, *args, @logger).run
    end

    def has_vm?(*args)
      Actions::HasVm.new(@vm_manager, *args, @logger).run
    end

    def reboot_vm(*args)
      Actions::RebootVm.new(@vm_manager, *args, {}, @logger).run
    end

    def set_vm_metadata(*args)
      Actions::SetVmMetadata.new(@vm_manager, *args, @logger).run
    end

    def configure_networks(*args)
      raise NotImplementedError, __method__
    end

    def create_disk(*args)
      Actions::CreateDisk.new(@disk_manager, *args, @logger).run
    end

    def delete_disk(*args)
      Actions::DeleteDisk.new(@disk_manager, *args, @logger).run
    end

    def attach_disk(*args)
      Actions::AttachDisk.new(@vm_manager, @disk_manager, *args, "persistent", @logger).run
    end

    def detach_disk(*args)
      Actions::DetachDisk.new(@vm_manager, @disk_manager, *args, "persistent", @logger).run
    end

    def has_disk?(disk_id)
      Actions::HasDisk.new(@disk_manager, disk_id, @logger).run
    end

    def get_disks(vm_id)
      Actions::GetDisks.new(@vm_manager, vm_id, @logger).run
    end

    def current_vm_id
      raise NotImplementedError, __method__
    end

    def snapshot_disk(disk_id, metadata={})
      raise NotImplementedError, __method__
    end

    def delete_snapshot(snapshot_id)
      raise NotImplementedError, __method__
    end
  end
end
