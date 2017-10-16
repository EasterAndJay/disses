require_relative 'worker_base'

class Worker < WorkerBase
  def initialize(node)
    super
    @node = node
  end

  def loop!
    if @node.tasks.empty?
      sleep 0.1
      return
    end

    message    = @node.tasks.pop
    @node.time = [@node.time, message.time].max + 1
    # q = @queue.map {|m| "(#{m.node}@#{m.time})"}
    # print "#{@node.port} @: #{message}  #{q.join(' ')}\n"
    print "#{@node.port} @: #{message}}\n"

    case(message.type)
    when :ENQUEUE
      @node.request!
    when :REQUEST
      @node.enqueue(message)
      @node.send(:ACKNOWLEDGE, message.node)
    when :ACKNOWLEDGE
      next if message < @myreq
      @node.await.delete(message.node)
    when :RELEASE
      @node.unqueue(message.node)
    end

    @node.like?
  end
end
