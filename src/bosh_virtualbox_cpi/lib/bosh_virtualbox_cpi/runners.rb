module BoshVirtualBoxCpi::Runners; end

runners_dir_path = File.expand_path("#{__FILE__}/../runners")
Dir["#{runners_dir_path}/*.rb"].each { |f| require(f) }
