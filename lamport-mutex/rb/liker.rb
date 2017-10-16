require 'set'
require 'socket'

require_relative 'message'
require_relative 'udpnetwork'

class Liker
  def initialize(port, others)
    @time   = 0
    @port   = port
    @others = others.reject {|p| p == @port}

    @tasks  = Queue.new

    @myreq  = nil
    @await  = others.to_set
    @queue  = Array.new

    @network = UDPNetwork.new(@tasks, @port)

    @worker = Worker.new do
      sleep 0.1 and next if @tasks.empty?
      self.handle(@tasks.pop)
    end

    @sender = Worker.new do
      if Random.rand < 0.02
        message = self.build(:ENQUEUE)
        @tasks.push(message)
      end

      sleep 0.1
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

    self.like?
  rescue => error
    puts "#{@port}: #{error}"
    puts error.backtrace
  end

  def like?
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

  def run!
    @network.start!
    @worker.start!
    @sender.start!
  end

  def send(message, *targets)
    unless message.is_a? Message
      message = self.build(message)
    end

    @network.send(message, *targets)
    return message
  end

  def stop!
    @network.stop!
    @worker.stop!
    @sender.stop!
  end

  def unqueue(pid)
    @queue.reject! do |queued|
      queued.pid == pid
    end
  end

  def wait!
    @sender.wait!
    @network.wait!
    @worker.wait!
  end
end
