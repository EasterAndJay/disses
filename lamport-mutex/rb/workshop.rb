require_relative 'worker'

class Workshop
  def initialize
    @workers = Array.new
  end

  def start!
    @workers.each(&:start!)
  end

  def stop!
    @workers.each(&:stop!)
  end

  def loop(&block)
    @workers << Worker.new(&block)
  end

  def wait!
    @workers.each(&:wait!)
  end
end
