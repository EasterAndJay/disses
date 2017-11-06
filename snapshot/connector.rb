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
    # Serve forever.
    Thread.new { listen }

    # Connect to lesser processes.
    threads = (1...@pid).map do |ppid|
      Thread.new { connect(ppid) }
    end

    # Wait for all connections.
    threads.each { |t| t.join }
    while @peers.length < @peer_count
      sleep rand
    end
  end

  def connect(peer_pid)
    peer = TCPSocket.open("localhost", BASE_PORT + peer_pid)
    peer.puts(@pid)
    add_peer(peer, peer_pid)
  rescue Errno::ECONNREFUSED
    @client.log "connection to client #{peer_pid} refused"
    sleep rand
    retry
  rescue Exception => e
    @client.log e
    sleep rand
    retry
  end

  def listen
    server = TCPServer.open("localhost", @port)
    loop do
      begin
        peer = server.accept_nonblock
        peer_pid = peer.gets.to_i
        if peer_pid == -42
          @client.initiate!
          peer.close
        else
          add_peer(peer, peer_pid)
        end
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
