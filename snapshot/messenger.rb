require_relative 'message_pb'

class Messenger

  def initialize(peers:)
    @peers = peers
  end

  def send_and_recv!
    recv_threads = @peers.map{ |pid, peer| Thread.new { recv_msgs(peer) }}
    send_threads = @peers.map{ |pid, peer| Thread.new { send_msgs(peer) }}
    recv_threads.each{ |t| t.join }
    send_threads.each{ |t| t.join }
  end

  def recv_msgs(peer)
    loop do
      data = peer.gets
      next if data.nil?
      msg = Message.decode_json(data)
      case msg.msg_type
      when Message::Type.resolve(:TRANSFER)
        handle_transfer
      when Message::Type.resolve(:MARKER)
        handle_marker
      end
    end
  end

  def handle_transfer
    # TODO
  end

  def handle_marker
    # TODO
  end

  def send_msgs(peer)
    loop do
      send_transfer(peer, 10) if rand <= 0.2
      sleep 10
    end
  end

  def send_message(peer, type, amount)
    msg = Message.new({
      msg_type: Message::Type.resolve(type),
      amount:   amount
    })
    peer.puts(msg.to_json)
  end

  def send_transfer(peer, amount)
    # TODO: Modify state of balance
    send_message(peer, :TRANSFER, amount)
  end

  def send_marker(peer)
    send_message(peer, :MARKER, 0)
  end


end