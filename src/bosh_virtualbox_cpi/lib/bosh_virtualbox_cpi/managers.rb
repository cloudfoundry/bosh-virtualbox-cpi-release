module BoshVirtualBoxCpi::Managers; end

managers_dir_path = File.expand_path("#{__FILE__}/../managers")
Dir["#{managers_dir_path}/*.rb"].each { |f| require(f) }
