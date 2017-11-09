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
    color   = @id.split('-').last.to_i(16) % 6 + 31
    result  = "==============================================\n"
    result << "SNAPSHOT: #{@id}\n"
    result << "CLIENT:   Client #{@client.pid}\n"
    result << "BALANCE:  \e[#{color}m$#{@balance} + $#{@incoming}\e[0m\n"

    unless @messages.empty?
      result << "MESSAGES:\n  - "
      result << @messages.join("\n  - ")
      result << "\n"
    end

    result << "\n"
  end
end
