require 'set'
require 'socket'

require_relative 'message'

class Liker
  def initialize(port, others)
    @time   = 0
    @port   = port
    @others = others.reject {|p| p == @port}
    @queue  = Array.new

    @tasks  = Queue.new

    @myreq  = nil
    @await  = others.to_set

    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', @port)

    @listener = WorkerBase.new do
      ...
    end


    @worklook = WorkerBase.new do
      #
    end
  end

  def build(type)
    Message.new({
      type: Message::Type.resolve(type),
      time: @time,
      node: @port
    })
  end

  def enqueue(message)
    @queue.push(message)
    @queue.sort!
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

  def listen!
    @listener = Thread.new do
      print "Starting receiver #{@port}...\n"
      while @running
        begin
          if select([@socket], nil, nil, 0.1)
            message = Message.decode_json(@socket.recv 1024)
            print "#{@port} <- #{message}\n"
            @tasks.push(message)
          else
            sleep 0.1
          end
        rescue => error
          print "#{@port}: #{error}\n#{error.backtrace}\n"
        end
      end
    end
  end

  def release!
    @myreq = nil
    self.unqueue(@port)
    self.send(:RELEASE, *@others)
  end

  def request!
    return unless @myreq.nil?
    @myreq = self.build(:REQUEST)
    @await = @others.to_set
    self.enqueue(@myreq)

    self.send(@myreq, *@others)
  end

  def run!
    @running = true
    self.listen!
    self.work!

    @sender = Thread.new do
      print "Starting sender #{@port}...\n"
      while @running
        begin
          if Random.rand < 0.02
            self.send(:ENQUEUE, @port)
          end

          sleep 0.1
        rescue => error
          print "#{@port}: #{error}\n#{error.backtrace}\n"
        end
      end
    end
  end

  def send(message, *targets)
    unless message.is_a? Message
      message = self.build(message)
    end

    @network.send(message)
    return message

    targets.each do |target|
      @socket.send(message.to_json, 0, '127.0.0.1', target)
      print "#{@port} -> [%11s  to  %4d at %4d]\n" % [
        message.type,
        message.node,
        message.time
      ]
    end
  end

  def stop!
    @running = false
  end

  def unqueue(node)
    @queue.reject! do |queued|
      queued.node == node
    end
  end

  def wait
    @sender.join   if @sender
    @listener.join if @listener
    @worker.join   if @worker
  end

  def work!
    @worker = Thread.new do
      while @running
        if @tasks.empty?
          sleep 0.1
          next
        end

        begin
          message = @tasks.pop
          @time   = [@time, message.time].max + 1

          q = @queue.map {|m| "(#{m.node}@#{m.time})"}
          print "#{@port} @: #{message}  #{q.join(' ')}\n"

          case(message.type)
          when :ENQUEUE
            self.request!
          when :REQUEST
            self.enqueue(message)
            self.send(:ACKNOWLEDGE, message.node)
          when :ACKNOWLEDGE
            next if message < @myreq
            @await.delete(message.node)
          when :RELEASE
            self.unqueue(message.node)
          end

          self.like?
        rescue => error
          puts "#{@port}: #{error}"
          puts error.backtrace
        end
      end
    end
  end
end
