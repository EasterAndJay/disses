require_relative 'message_pb'

class Message
  include Comparable

  def to_s
    "[%11s from %4d at %4d]" % [
      self.msg_type,
      self.pid,
      self.clock
    ]
  end

  def <=> (other)
    if self.clock == other.clock
      self.pid <=> other.pid
    else
      self.clock <=> other.clock
    end
  end
end
