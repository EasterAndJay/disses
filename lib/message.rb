require_relative 'messages_pb'

class Message
  include Comparable

  def to_s
    "[%11s from %4d at %4d]" % [
      self.type,
      self.node,
      self.time
    ]
  end

  def <=> (other)
    if self.time == other.time
      self.node <=> other.node
    else
      self.time <=> other.time
    end
  end
end
