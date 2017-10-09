class Worker
  def initialize(liker, queue)
    @liker  = liker
    @queue  = queue
    @thread = nil
  end

  def start!
    @active   = true
    @thread ||= Thread.new do
      while @active
        begin
          message = @queue.pop

          case message.type
          when 'request'
            liker.enqueue(message)
            liker.send('grant', message.node)
          when 'grant'
            next if message < @myreq
            @await.delete(message.node)
            liker.like?
          when 'release'
            liker.unqueue(message.node)
            liker.like?
          end
        rescue IO::WaitReadable
          # Everything is fine.
        rescue => error
          puts "#{@port}: #{error}"
          puts error.backtrace
        end
      end
    end
  end

  def stop!
    @active = false
  end
end
