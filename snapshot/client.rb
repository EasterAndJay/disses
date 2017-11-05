require 'thread'

require_relative './connector'
require_relative './messenger'

class Client

  attr_reader :pid

  def initialize(pid:, network_size:)
    @pid = pid
    @connector = Connector.new(self, peer_count: network_size - 1)
    @messenger = nil

    @balance_lock = Mutex.new
    @balance = 1000

    @marker_signal = ConditionVariable.new
    @markers = 0
  end

  def log(message)
    if message.is_a? Exception
      print "Client #{@pid}:  #{message}\n  #{message.backtrace.join("\n  ")}\n"
    else
      print "Client #{@pid}:  #{message}\n"
    end
  end

  def run!
    @connector.init_connections!
    log "connected to all other peers"

    @messenger = Messenger.new(self, peers: peers)
    Thread.new{ @messenger.send_and_recv! }
    loop do
      snapshot! if gets.chomp == "snapshot"
    end
  end

  def snapshot!
    log "nitiating snapshot"

    @balance_lock.synchronize {
      state = @balance
      peers.each do |pid, peer|
        @messenger.send_marker(peer)
        # TODO: Save state on every channel
      end
      @marker_signal.wait(@balance_lock)
    }
    p "Client #{@connector.pid}: State of snapshot"
    p "Balance = $#{@balance}"
    p "Channels: #{@channel_states}"
  end

  private

  def peers
    @connector.peers
  end
end
