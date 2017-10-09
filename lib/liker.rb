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

    @tasks  = Queue.new

    @myreq  = nil
    @await  = nil

    @socket = UDPSocket.new
    @socket.bind('127.0.0.1', @port)
  end

  def enqueue(message)
    @queue.push(message)
    @queue.sort! do |a, b|
      (a.time == b.time)? a.node <=> b.node : a.time <=> b.time
    end

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
    @listener = Thread.new do
      print "Starting receiver #{@port}...\n"
      while @running
        begin
          if select([@socket], nil, nil, 0.1)
            message = Message.parse(@socker.recv)
            print "#{@port} <- #{message}\n"
            @tasks.push(message)
          else
            sleep 0.1
            next
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
    self.work!

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
      message = Request.new(type: 'huh', time: @time, node: @port)
      # message = Message.new(type, @time, @port)
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

          case(message.type)
          when ''
          when 'request'
            self.enqueue(message)
            self.send('grant', message.node)
          when 'grant'
            next if message < @myreq
            @await.delete(message.node)
            self.like?
          when 'release'
            self.unqueue(message.node)
            self.like?
          end
        rescue => error
          puts "#{@port}: #{error}"
          puts error.backtrace
        end
      end
    end
  end
end
