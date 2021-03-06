require 'logger'
require 'securerandom'
require 'thread'

require_relative './connector'
require_relative './messenger'
require_relative './snapshot'

class Client

  attr_reader :balance
  attr_reader :log
  attr_reader :pid
  attr_reader :tasks

  def initialize(pid:, network_size:, auto: false, verbosity: Logger::INFO)
    @pid  = pid
    @auto = auto

    @log = Logger.new(STDOUT)
    @log.level = verbosity
    @log.formatter = proc do |severity, datetime, progname, msg|
      if msg.is_a? Exception
        "Client #{@pid}:  #{msg}\n  #{msg.backtrace.join("\n  ")}\n"
      else
        "Client #{@pid}:  #{msg}\n"
      end
    end

    @connector = Connector.new(self, peer_count: network_size - 1)
    @messenger = nil

    @balance_lock = Mutex.new
    @balance = 1000

    @tasks = Queue.new
    @snaps = Hash.new

    Thread.new do
      loop {handle @tasks.pop}
    end
  end

  def handle(message)
    case message.msg_type
    when :MARKER
      self.snapshot! message
    when :TRANSFER
      self.rebalance! message
    else
      self.log.error "unknown message type: '#{message.msg_type}'"
    end

    @snaps.reject! do |id, snap|
      snap.handle(message)
      print snap if snap.done?
      snap.done?
    end
  rescue Exception => e
    self.log.error e
  end

  def initiate!
    self.tasks.push Message.new({
      msg_type: Message::Type.resolve(:MARKER),
      ssid: SecureRandom.uuid,
      ppid: self.pid
    })
  end

  def rebalance!(message)
    @balance_lock.synchronize do
      self.log.debug "got $#{message.amount} from client #{message.ppid}"
      self.log.info  "now has $#{@balance} + $#{message.amount} = $#{@balance + message.amount}"
      @balance += message.amount
    end
  end

  def run!
    @connector.init_connections!
    self.log.info "connected to all other peers"

    @messenger = Messenger.new(self, peers: peers)
    Thread.new { @messenger.send_and_recv! }

    loop do
      sleep rand * 30
      self.initiate! if @auto
    end
  end

  def snapshot!(message)
    if @snaps.include? message.ssid
      # Already working on this one...
      return
    end

    @balance_lock.synchronize do
      self.log.info "taking snapshot #{message.ssid}"
      @snaps[message.ssid] = Snapshot.new(self, message, peers.keys)
      sleep rand # Let some messages build up!

      peers.each do |pid, peer|
        peer.print(Message.new({
          msg_type: Message::Type.resolve(:MARKER),
          ssid: message.ssid,
          ppid: self.pid
        }).to_json << "\n")
      end
    end
  end

  def transfer!(pid)
    amount = rand(9) + 1
    @balance_lock.synchronize do
      @balance -= amount
    end

    self.log.debug "sending $#{amount} to client #{pid}"
    peers[pid].print(Message.new({
      msg_type: Message::Type.resolve(:TRANSFER),
      amount:   amount,
      ppid:     self.pid
    }).to_json << "\n")
  rescue Exception => e
    self.log.error e
  end

  private

  def peers
    @connector.peers
  end
end
