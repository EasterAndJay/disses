# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: message.proto

require 'google/protobuf'

Google::Protobuf::DescriptorPool.generated_pool.build do
  add_message "Message" do
    optional :msg_type, :enum, 1, "Message.Type"
    optional :amount, :uint32, 2
    optional :ppid, :uint32, 3
    optional :ssid, :string, 4
  end
  add_enum "Message.Type" do
    value :TRANSFER, 0
    value :MARKER, 1
  end
end

Message = Google::Protobuf::DescriptorPool.generated_pool.lookup("Message").msgclass
Message::Type = Google::Protobuf::DescriptorPool.generated_pool.lookup("Message.Type").enummodule
