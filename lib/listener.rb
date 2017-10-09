class Listener
  def initialize(queue, port)
    @queue  = queue
    @thread = nil

    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', @port)
  end

  def send(*targets)
    targets.each do |target|
      message = Message.new(type, @time, @port)
      print "#{@port} -> #{type} to #{target} at #{@time}\n"
      @socket.send(message.to_json, 0, '127.0.0.1', target)
    end
  end

  def start!
    @active   = true
    @thread ||= Thread.new do
      while @active
        begin
          receive = @socket.recvfrom_nonblock(1024)
          next if receive.first.empty?
          print "#{@port} <- #{message}\n"
          @queue.push(decode(message))
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
