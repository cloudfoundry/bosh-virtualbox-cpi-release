module BoshVirtualBoxCpi
  class ResolvedDir
    def initialize(path, runner)
      @path = path
      @runner = runner
    end

    def to_s
      @resolved_path ||= if @path.start_with?("~/")
        home_path = @runner.execute!("bash", "-c", "echo $HOME")[1].strip
        raise "Home path must not be empty" if home_path.empty?
        home_path + @path[1..-1] + "/"
      else
        @path + "/"
      end
    end
  end
end
