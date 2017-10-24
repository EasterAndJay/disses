require 'thread'

require_relative './connector'
require_relative './messenger'

class Client

  def initialize(pid:, network_size:)
    @pid = pid
    @peer_count = network_size - 1
    @peers = Hash.new
    @connector = Connector.new(peers: @peers, pid: @pid, peer_count: @peer_count)
    @messenger = nil

    @balance_lock = Mutex.new
    @balance = 1000
    @marker_signal = ConditionVariable.new
    @markers = 0
    @track_channel_state = false
    @channel_states = Hash.new
  end

  def run!
    @connector.init_connections!
    p "Client #{@pid}: Connected to all other peers"
    @messenger = Messenger.new(peers: @peers)
    @messenger.send_and_recv!
    loop do
      snapshot! if gets.chomp == "snapshot"
    end
  end

  def snapshot!
    @balance_lock.synchronize {
      state = @balance
      @peers.each do |pid, peer|
        @messenger.send_marker(peer)
        # TODO: Save state on every channel
      end
      @marker_signal.wait(@balance_lock)
    }
    p "Client #{@pid}: State of snapshot"
    p "Balance = $#{@balance}"
    p "Channels: #{@channel_states}"
  end
end