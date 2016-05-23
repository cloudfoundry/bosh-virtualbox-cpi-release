module BoshVirtualBoxCpi::Virtualbox; end

virtualbox_dir_path = File.expand_path("#{__FILE__}/../virtualbox")
Dir["#{virtualbox_dir_path}/*.rb"].each { |f| require(f) }
