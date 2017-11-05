require_relative 'message_pb'

class Messenger

  def initialize(client, peers:)
    @client = client
    @peers  = peers
  end

  def send_and_recv!
    recv_threads = @peers.map{ |pid, peer| Thread.new { recv_msgs(pid, peer) }}
    send_threads = @peers.map{ |pid, peer| Thread.new { send_msgs(pid, peer) }}
    recv_threads.each{ |t| t.join }
    send_threads.each{ |t| t.join }
  end

  def recv_msgs(pid, peer)
    loop do
      begin
        data = peer.gets
        next if data.nil?

        msg = Message.decode_json(data)
        @client.tasks.push msg
      rescue Exception => e
        @client.log e
      end
    end
  end

  def send_msgs(pid, peer)
    loop do
      @client.transfer!(pid)
      sleep rand * 5
    end
  end

end
