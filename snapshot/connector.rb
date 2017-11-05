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
    threads << Thread.new { listen }
    threads << Thread.new do
      (0...@pid).each do |peer_pid|
        connect(peer_pid)
      end
    end

    threads.each { |t| t.join }
  end

  def connect(peer_pid)
    peer = TCPSocket.open("localhost", BASE_PORT + peer_pid)
    peer.puts(@pid)
    add_peer(peer, peer_pid)
  rescue Errno::ECONNREFUSED
    sleep rand
  rescue Exception => e
    @client.log e
    sleep rand
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
  ensure
    server.close
  end

  def add_peer(peer, peer_pid)
    if @peers.key?(peer_pid)
      @client.log "already connected to client #{peer_pid}"
      peer.close
    else
      @client.log "connected to client #{peer_pid}"
      @peers[peer_pid] = peer
    end
  end

end
