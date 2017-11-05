require 'socket'
require 'thread'
require 'concurrent'

BASE_PORT = 5000

class Connector

  attr_reader :pid, :peers

  def initialize(client, peer_count:)
    @client = client
    @peers = Concurrent::Hash.new
    @peer_count = peer_count

    @pid  = client.pid
    @port = @pid + BASE_PORT
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
      rescue Errno::ECONNREFUSED
        sleep rand
      rescue Exception => e
        @client.log e
        sleep rand
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
      rescue Errno::EWOULDBLOCK
        # Everything is fine.
        sleep rand
      rescue Exception => e
        @client.log e
      end
    end
    server.close
  end

  def add_peer(peer, peer_pid)
    if @peers.length < @peer_count && !@peers.key?(peer_pid)
      @client.log "connected to client #{peer_pid}"
      @peers[peer_pid] = peer
    else
      peer.close
    end
  end

end
