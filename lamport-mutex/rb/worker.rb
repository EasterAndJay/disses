class Worker
  def initialize(&block)
    @block  = block
    @active = false
    @thread = nil
  end

  def start!
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

  def stop!
    @active = false
  end

  def wait!
    @thread.join if @thread
  end
end