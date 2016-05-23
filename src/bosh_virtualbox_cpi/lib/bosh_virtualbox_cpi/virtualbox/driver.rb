require "shellwords"
require "bosh_virtualbox_cpi/virtualbox/error"
require "bosh_virtualbox_cpi/virtualbox/vm_importer"
require "bosh_virtualbox_cpi/virtualbox/vm_cloner"
require "bosh_virtualbox_cpi/virtualbox/vm_finder"
require "bosh_virtualbox_cpi/virtualbox/cdrom_mounter"
require "bosh_virtualbox_cpi/virtualbox/network_configurer"
require "bosh_virtualbox_cpi/virtualbox/disk_attacher"
require "bosh_virtualbox_cpi/virtualbox/disk_creator"
require "bosh_virtualbox_cpi/virtualbox/resume_pause_hot_plugger"

module BoshVirtualBoxCpi::Virtualbox
  class Driver
    def initialize(runner, bin_path, retrier, logger)
      @runner = runner
      @bin_path = bin_path
      @retrier = retrier
      @logger = logger
    end

    def execute(*args)
      options = args.last.is_a?(Hash) ? args.pop : {}

      # Last exit code and output
      exit_code, output = nil, nil

      @retrier.times do
        exit_code, output = @runner.execute(@bin_path, *args) # todo path

        # Typically happens when running showvminfo
        if exit_code != 0 && output =~ /VBoxManage: error: The object is not ready/
          raise BoshVirtualBoxCpi::Virtualbox::Error, "Retrying after 'object is not ready' error"
        end
      end

      if exit_code != 0
        if exit_code == 126
          # This exit code happens if VBoxManage is on the PATH,
          # but another executable it tries to execute is missing.
          # This is usually indicative of a corrupted VirtualBox install.
          raise BoshVirtualBoxCpi::Virtualbox::Error, "Most likely corrupted VirtualBox installation"
        else
          errored = !options[:ignore_non_zero_exit_code]
        end
      else
        # Sometimes, VBoxManage fails but doesn't actual return a non-zero exit code.
        if output =~ /failed to open \/dev\/vboxnetctl/i
          # This catches an error message that only shows when kernel
          # drivers aren't properly installed.
          raise BoshVirtualBoxCpi::Virtualbox::Error, "Error message about vboxnetctl"
        end

        if output =~ /VBoxManage: error:/
          @logger.info("VBoxManage error text found, assuming error.")
          errored = true
        end
      end

      if errored
        raise BoshVirtualBoxCpi::Virtualbox::Error, <<-MSG
          Error executing command:
            Command:   '#{args.inspect}'
            Exit code: '#{exit_code}'
            Output:    '#{output}'
        MSG
      end

      output.gsub("\r\n", "\n")
    end

    def vm_importer
      VmImporter.new(self, @retrier, @logger)
    end

    def vm_cloner
      VmCloner.new(self, @logger)
    end

    def vm_finder
      VmFinder.new(self, @logger)
    end

    def cdrom_mounter(vm)
      CdromMounter.new(self, vm, resume_pause_hot_plugger(vm), @logger)
    end

    def network_configurer(vm)
      NetworkConfigurer.new(self, vm, @logger)
    end

    def disk_creator
      DiskCreator.new(self, @logger)
    end

    def disk_attacher(vm)
      DiskAttacher.new(self, vm, resume_pause_hot_plugger(vm), @logger)
    end

    private

    def resume_pause_hot_plugger(vm)
      ResumePauseHotPlugger.new(self, vm, @logger)
    end
  end
end
