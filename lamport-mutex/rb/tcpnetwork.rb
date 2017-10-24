require_relative 'message'

class TCPNetwork
  def initialize(queue, port)
    @socket  = TCPServer.new('127.0.0.1', port)
    @clients = Hash.new
    @queue   = queue
    @port    = port
  end

  def client(port)
    @clients[port] ||= begin
      client = TCPSocket.open('127.0.0.1', port)
      client.puts @port
      client
    end
  rescue
    sleep 0.1
    retry
  end

  def serve
    client = @socket.accept_nonblock
    port   = client.gets.to_i
    puts "Connection from #{port}"
    if @clients.include? port
      print "#{@port} XX Already connected to #{port}.\n"
      client.close
    else
      @clients[port] = client
    end
  rescue IO::WaitReadable
    # Everything is fine.
    sleep 0.1
  end

  def recv
    clients = select(@clients.values, nil, nil, 0.1)
    return if clients.nil?

    clients[0].each do |client|
      data = client.gets
      next if data.nil?

      message = Message.decode_json(data)
      print "#{@port} <- #{message}\n"
      @queue.push(message)
    end
  end

  def send(message, *targets)
    targets.each do |target|
      client = self.client(target)
      return if client.nil?

      client.puts(message.to_json)
      print("%4d -> [%11s  to  %4d at %4d]\n" % [
        message.pid,
        message.msg_type,
        target,
        message.clock
      ])
    end
  end
end
