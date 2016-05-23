#!/usr/bin/env ruby

require "rubydns"

# Sample manifests specify that host (192.168.56.1) is a DNS server.
# Easiest way to satisfy that is to spin up a local DNS server
# that passes through requests to the internet.
# `sudo ./local_dns_server.rb`

# Use upstream DNS for name resolution.
UPSTREAM = RubyDNS::Resolver.new([[:udp, "8.8.8.8", 53], [:tcp, "8.8.8.8", 53]])

RubyDNS.run_server(listen: [[:udp, "0.0.0.0", 53], [:tcp, "0.0.0.0", 53]]) do
  otherwise do |transaction|
    transaction.passthrough!(UPSTREAM)
  end
end
