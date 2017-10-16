require_relative 'message'
require_relative 'worker_base'

class UDPNetwork < Worker
  def initialize(node, port)
    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', @node.port)

    super do
      if select([@socket], nil, nil, 0.1)
        message = Message.decode_json(@socket.recv(1024))
        # print "#{@port} <- #{message}\n"
        @node.tasks.push(message)
      else
        sleep 0.1
      end
    end
  end

  def send(message, *targets)
    targets.each do |target|
      @socket.send(message.to_json, 0, '127.0.0.1', target)
      # print("#{@port} -> [%11s  to  %4d at %4d]\n" % [
      #   message.type,
      #   message.node,
      #   message.time
      # ])
    end
  end
end
