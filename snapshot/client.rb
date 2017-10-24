require_relative './connector'

class Client

  def initialize(pid:, network_size:)
    @pid = pid
    @peer_count = network_size - 1
    @balance = 1000
    @peers = Hash.new
    @connector = Connector.new(peers: @peers, pid: @pid, peer_count: @peer_count)
  end

  def run!
    @connector.init_connections!
    p "Client #{@pid}: Connected to all other peers"
  end

end