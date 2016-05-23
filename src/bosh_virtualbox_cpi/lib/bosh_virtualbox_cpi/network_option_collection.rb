require "bosh_virtualbox_cpi/network_option"

module BoshVirtualBoxCpi
  class NetworkOptionCollection
    def self.from_hash(hash)
      new(hash.map do |name, properties|
        NetworkOption.new.tap do |no|
          no.name    = name
          no.ip      = properties["ip"]
          no.netmask = properties["netmask"]
          no.gateway = properties["gateway"]
          no.dns     = properties["dns"]
          no.default = properties["default"]
          no.cloud   = properties["cloud_properties"]
        end
      end)
    end

    def initialize(network_options)
      @network_options = network_options
    end

    include Enumerable
    def each(*args, &blk)
      @network_options.each(*args, &blk)
    end
  end
end
