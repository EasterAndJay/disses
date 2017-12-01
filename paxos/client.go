package main

import (
  "github.com/golang/protobuf/proto";

  "fmt";
  "net";
)

type Client struct {
  port int;
  work chan *Message;
}

func (client *Client) handle() {
  for {
    message := <-client.work

    switch message.GetType() {
    case Message_S_PROPOSE:
      fmt.Printf("Got propositioned\n")
    case Message_R_PROPOSE:
      fmt.Printf("Got a prop response\n")
    case Message_S_ACCEPT:
      fmt.Printf("Got sent a value\n")
    case Message_R_ACCEPT:
      fmt.Printf("Got a value response\n")
    case Message_S_NOTIFY:
      fmt.Printf("Got sent an update\n")
    default:
      fmt.Printf("The hell is that?\n")
    }
  }
}

func (client *Client) listen() {
  buffer := make([]byte, 2048)
  server, err := net.ListenUDP("udp", &net.UDPAddr {
    IP:   net.ParseIP("127.0.0.1"),
    Port: client.port,
  })

  if err != nil {
    fmt.Printf("Listen error: %v\n", err)
    return
  }

  defer server.Close()

  for {
    _, _, err := server.ReadFromUDP(buffer)
    if err != nil {
      fmt.Printf("Receive error: %v\n", err)
      continue
    }

    message := new(Message)
    err = proto.Unmarshal(buffer, message)
    if err != nil {
      fmt.Printf("Parse error: %v\n", err)
      continue
    }

    client.work <- message
  }
}

func NewClient(port int) Client {
  return Client{port, make(chan *Message, 16)}
}
