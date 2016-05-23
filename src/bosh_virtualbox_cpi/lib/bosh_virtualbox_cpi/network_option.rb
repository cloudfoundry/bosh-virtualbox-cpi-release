module BoshVirtualBoxCpi
  class NetworkOption
    attr_accessor(
      :name,
      :ip,
      :netmask,
      :gateway,
      :dns,
      :default,
      :mac,
      :cloud,
    )

    def formatted_mac
      mac.downcase.scan(/../).join(":")
    end

    def cloud_name
      cloud["name"] || "vboxnet0"
    end

    def cloud_type
      cloud["type"] || "hostonly"
    end
  end
end
