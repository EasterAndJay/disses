class Worker
  def initialize(&block)
    @children = Array.new
    @block    = block
    @active   = false
    @thread   = nil
  end

  def start!
    @children.each(&:start!)

    if @block
      @active = true
      @thread ||= Thread.new do
        while @active
          begin
            @block.call
          rescue => error
            print("#{error}\n#{error.backtrace}\n")
          end
        end

        @thread = nil
      end
    end
  end

  def stop!
    @children.each(&:stop!)
    @active = false
  end

  def subtask(worker = nil, &block)
    if block
      @children << Worker.new(&block)
    elsif worker.is_a? Worker
      @children << worker
    else
      raise "What is that?"
    end
  end

  def wait!
    @children.each(&:wait!)
    @thread.join if @thread
  end
end
