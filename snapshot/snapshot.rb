require 'set'

class Snapshot
  def initialize(client, message, pids)
    @client = client
    @id     = message.ssid
    @pids   = pids.to_set

    @balance  = @client.balance
    @incoming = 0
    @messages = Array.new
  end

  def done?
    @pids.empty?
  end

  def handle(message)
    unless @pids.include? message.ppid
      # Already got a marker from you.
      return
    end

    case message.msg_type
    when :TRANSFER
      @messages << "$#{message.amount} from client #{message.ppid}"
      @incoming += message.amount
    when :MARKER
      return unless @id == message.ssid
      @pids.delete message.ppid
    end
  end

  def to_s
    <<~EOF
      CLIENT:   Client #{@client.pid}
      SNAPSHOT: #{@id}
      BALANCE:  $#{@balance} + $#{@incoming}
      MESSAGES:
        - #{@messages.join("\n  - ")}
    EOF
  end
end
