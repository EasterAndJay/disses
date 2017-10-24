require_relative 'message'

class UDPNetwork
  def initialize(queue, port)
    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', port)
    @queue = queue
    @port  = port
  end

  def recv
    return unless select([@socket], nil, nil, 0.1)
    message = Message.decode_json(@socket.recv(1024))
    print "#{@port} <- #{message}\n"
    @queue.push(message)
  end

  def send(message, *targets)
    targets.each do |target|
      @socket.send(message.to_json, 0, '127.0.0.1', target)
      print("%4d -> [%11s  to  %4d at %4d]\n" % [
        message.pid,
        message.msg_type,
        target,
        message.clock
      ])
    end
  end
end
