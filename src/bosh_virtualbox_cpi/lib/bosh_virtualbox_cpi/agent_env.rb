require "json"

module BoshVirtualBoxCpi
  class AgentEnv
    attr_accessor *%w(
      vm_id
      name
      agent_id
      networks
      disks
      env
      agent_env
    )

    def self.from_json(json)
      from_hash(JSON.parse(json))
    end

    def self.from_hash(hash)
      hash = hash.clone

      new.tap do |ae|
        hash.delete("vm").tap do |vm_hash|
          ae.vm_id = vm_hash.delete("id")
          ae.name  = vm_hash.delete("name")
        end

        ae.agent_id = hash.delete("agent_id")
        ae.networks = hash.delete("networks")
        ae.disks    = hash.delete("disks")
        ae.env      = hash.delete("env")
        ae.agent_env = hash
      end
    end

    def add_networks(network_options)
      @networks = Hash[network_options.map do |no|
        [no.name, {
          "ip"      => no.ip,
          "netmask" => no.netmask,
          "gateway" => no.gateway,
          "dns"     => no.dns,
          "default" => no.default || [], # bosh_agent requires it to be an array or no key
          "mac"     => no.formatted_mac,
          "cloud_properties" => no.cloud,
        }]
      end]
    end

    def add_empty_disks
      @disks = {
        "system" => "0",
        "ephemeral" => nil,
        "persistent" => {}
      }
    end

    def add_ephemeral_disk(disk_id, unit_number)
      raise "ephemeral disk is already set" \
        if @disks["ephemeral"]
      @disks["ephemeral"] = unit_number.to_s
    end

    def remove_ephemeral_disk(disk_id)
      @disks["ephemeral"] = nil
    end

    def add_persistent_disk(disk_id, unit_number)
      @disks["persistent"][disk_id] = unit_number
    end

    def remove_persistent_disk(disk_id)
      @disks["persistent"].delete(disk_id)
    end

    def as_hash
      {
        "vm" => {
          "id"   => @vm_id || raise("missing vm_id"),
          "name" => @name  || raise("missing name"),
        },
        "agent_id" => @agent_id || raise("missing agent_id"),
        "networks" => @networks || raise("missing networks"),
        "disks"    => @disks    || raise("missing disks"),
        "env"      => @env,
      }.merge(@agent_env || {})
    end

    def as_json
      JSON.dump(as_hash)
    end

    def as_iso
      Dir.mktmpdir do |dir|
        env_path = File.join(dir, "env")
        iso_path = File.join(dir, "env.iso")

        exe = "mkisofs"
        if File.exists?("/var/vcap/packages/genisoimage/genisoimage")
          exe = "/var/vcap/packages/genisoimage/genisoimage"
        end

        File.open(env_path, "w") { |f| f.write(as_json) }
        output = `#{exe} -o #{iso_path} #{env_path} 2>&1`
        raise "#{$?.exitstatus} - #{output}" unless $?.exitstatus.zero?

        File.open(iso_path, "r") { |f| f.read }
      end
    end
  end
end
