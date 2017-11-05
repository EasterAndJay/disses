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
        case msg.msg_type
        when :TRANSFER
          @client.log "got $#{msg.amount} from client #{pid}"
          handle_transfer(msg.amount)
        when :MARKER
          @client.log "got a MARKER from client #{pid}"
          handle_marker(pid, peer)
        end
      rescue Exception => e
        @client.log e
      end
    end
  end

  def handle_transfer(amount)
    # TODO: Modify state of balance
  end

  def handle_marker(pid, peer)
    # TODO
  end

  def send_msgs(pid, peer)
    loop do
      begin
        amount = rand(9) + 1
        @client.log "sending $#{amount} to client #{pid}"
        send_transfer(peer, amount)
      rescue Exception => e
        @client.log e
      end

      sleep rand * 5
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
