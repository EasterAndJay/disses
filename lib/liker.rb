require 'json'
require 'queue'
require 'set'
require 'socket'

require_relative 'message'

class Liker
  def initialize(port, others)
    @time   = 0
    @port   = port
    @others = others.reject {|p| p == @port}
    @queue  = Array.new

    @myreq  = nil
    @await  = nil

    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', @port)
  end

  def enqueue(message)
    @queue.push(message)
    @queue.sort!

    # print "#{@port}: #{@queue.inspect}\n"
  end

  def like?
    if @queue.first == @myreq and @await.empty?
      self.like! Random.rand(5) + 1
    end
  end

  def like!(amount = 1)
    score = File.read('likes.int').to_i
    print "\e[32m#{@port} likes it this much: #{score}+#{amount}\e[0m\n"
    File.write('likes.int', "#{score + amount}\n")
    sleep(Random.rand)
    print "\e[31m#{@port} is done\e[0m\n"
  ensure
    self.release!
  end

  def listen!
    @receiver = Thread.new do
      print "Starting receiver #{@port}...\n"
      while @running
        begin
          receive = @socket.recvfrom_nonblock(1024)
          next if receive.first.empty?

          message = Message.parse(receive.first)
          @time   = [@time, message.time].max + 1
          print "#{@port} <- #{message}\n"

          case message.type
          when 'request'
            self.enqueue(message)
            self.send('grant', message.from)
          when 'grant'
            next if message < @myreq
            @await.delete(message.from)
            self.like?
          when 'release'
            self.unqueue(message.from)
            self.like?
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

  def release!
    @myreq = nil
    self.unqueue(@port)
    self.send('release', *@others)
  end

  def request!
    return unless @myreq.nil?
    @myreq = Message.new('request', @time, @port)
    @await = @others.to_set
    self.enqueue(@myreq)

    self.send('request', *@others)
  end

  def run!
    @running = true
    self.listen!

    @sender = Thread.new do
      print "Starting sender #{@port}...\n"
      while @running
        begin
          sleep(Random.rand * 5)
          self.request!
        rescue => error
          puts "#{@port}: #{error}"
          puts error.backtrace
        end
      end
    end
  end

  def send(type, *targets)
    targets.each do |target|
      message = Message.new(type, @time, @port)
      print "#{@port} -> #{type} to #{target} at #{@time}\n"
      @socket.send(message.to_json, 0, '127.0.0.1', target)
    end
  end

  def stop!
    @running = false
  end

  def unqueue(from)
    @queue.reject! do |queued|
      queued.from == from
    end
  end

  def wait
    @sender.join   if @sender
    @receiver.join if @receiver
  end
end
