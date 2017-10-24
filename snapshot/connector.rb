require 'socket'
require 'thread'

BASE_PORT = 5000

class Connector

  def initialize(peers:, pid:, peer_count:)
    @peers_lock = Mutex.new
    @peers = peers
    @peer_count = peer_count

    @pid = pid
    @port = pid + BASE_PORT
  end

  def init_connections!
    threads = []
    threads << Thread.new { connect }
    threads << Thread.new { listen }
    threads.each { |t| t.join }
  end

  def connect
    peer_pid = -1
    while @peers.length < @peer_count
      peer_pid += 1
      peer_pid = 0 if peer_pid > @peer_count
      next if peer_pid == @pid
      peer_port = peer_pid + BASE_PORT
      begin
        peer = TCPSocket.open("localhost", peer_port)
        peer.puts(@pid)
        add_peer(peer, peer_pid)
      rescue Exception => e
        sleep 1
      end
    end
  end

  def listen
    server = TCPServer.open("localhost", @port)
    while @peers.length < @peer_count
      begin
        peer = server.accept_nonblock
        peer_pid = peer.gets.to_i
        add_peer(peer, peer_pid)
      rescue
      end
    end
    server.close
  end

  def add_peer(peer, peer_pid)
    @peers_lock.synchronize do
      if @peers.length < @peer_count && !@peers.key?(peer_pid)
        p "Client #{@pid}: Connected to client #{peer_pid}"
        @peers[peer_pid] = peer
      else
        peer.close
      end
    end
  end

end