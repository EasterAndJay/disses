require 'set'
require 'socket'

require_relative 'message'
require_relative 'tcpnetwork'
require_relative 'udpnetwork'
require_relative 'workshop'

class Liker
  def initialize(port, others, options = {})
    @time   = 0
    @port   = port
    @others = others.reject {|p| p == @port}

    @myreq  = nil
    @await  = others.to_set
    @queue  = Array.new


    @tasks   = Queue.new
    @workers = Workshop.new

    case(options[:protocol])
    when :tcp
      @network = TCPNetwork.new(@tasks, @port)
      @workers.loop {@network.serve}
      @workers.loop {@network.recv}
    when :udp
      @network = UDPNetwork.new(@tasks, @port)
      @workers.loop {@network.recv}
    else
      raise 'Unknown protocol!'
    end

    @workers.loop do
      sleep 0.1 and next if @tasks.empty?
      self.handle(@tasks.pop)
    end

    unless options[:manual]
      @workers.loop do
        if Random.rand < 0.02
          message = self.build(:ENQUEUE)
          @tasks.push(message)
        end

        sleep 0.1
      end
    end
  end

  def build(type)
    Message.new({
      msg_type: Message::Type.resolve(type),
      clock:    @time,
      pid:      @port
    })
  end

  def enqueue(message)
    @queue.push(message)
    @queue.sort!
  end

  def handle(message)
    @time = [@time, message.clock].max + 1
    q = @queue.map {|m| "(#{m.pid}@#{m.clock})"}
    print "#{@port} @: #{message}  #{q.join(' ')}\n"

    case(message.msg_type)
    when :ENQUEUE
      self.request!
    when :REQUEST
      self.enqueue(message)
      self.send(:ACKNOWLEDGE, message.pid)
    when :ACKNOWLEDGE
      return if message < @myreq
      @await.delete(message.pid)
    when :RELEASE
      self.unqueue(message.pid)
    end

    # If we control the mutex, go do stuff!
    if @myreq and @queue.first == @myreq and @await.empty?
      self.like! Random.rand(5) + 1
    end
  end

  def like!(amount = 1)
    score = File.read('likes.int').to_i
    print "\e[32m#{@port} likes it this much: #{score}+#{amount}\e[0m\n"
    sleep(Random.rand)
    File.write('likes.int', "#{score + amount}\n")
    print "\e[31m#{@port} is done\e[0m\n"
  ensure
    self.release!
  end

  def release!
    @myreq = nil
    self.unqueue(@port)
    self.send(:RELEASE, *@others)
  end

  def request!
    return unless @myreq.nil?
    @myreq = self.send(:REQUEST, *@others)
    @await = @others.to_set
    self.enqueue(@myreq)
  end

  def send(type, *targets)
    message = self.build(type)
    @network.send(message, *targets)
    return message
  end

  def start!
    @workers.start!
  end

  def stop!
    @workers.stop!
  end

  def unqueue(pid)
    @queue.reject! do |queued|
      queued.pid == pid
    end
  end

  def wait!
    @workers.wait!
  end
end
