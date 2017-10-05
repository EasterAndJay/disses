require 'json'

class Message
  include Comparable
  attr_reader :from
  attr_reader :time
  attr_reader :type

  def self.parse(json)
    data = JSON.parse(json)
    self.new(data['type'], data['time'], data['from'])
  end

  def initialize(type, time, from)
    @type = type
    @time = time
    @from = from
  end

  def to_json
    JSON.generate({
      type: @type,
      time: @time,
      from: @from
    })
  end

  def to_s
    "#{@type} from #{@from} at #{@time}"
  end

  def <=> (other)
    return @from <=> other.from if @time == other.time
    return @time <=> other.time
  end
end
