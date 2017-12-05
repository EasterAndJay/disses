package main

import "container/list"
import "fmt"
import "github.com/golang/protobuf/proto"
import "net"
import "time"

type Client struct {
  port int;
  sock *net.UDPConn;
  work chan *Message;
  logs list.List;
}

func (client *Client) Handle() {
  for {
    message := <-client.work

    epoch = message.GetEpoch()
    if epoch < logs.Len() {
      logs.Get(epoch).handle(message)
    }
    else if epoch == logs.Len() {
      logs.PushBack(NewEntry(message))
    }
    else {
      // Way in the future!
    }


    
  }
}

func (client *Client) Listen() {
  buffer := make([]byte, 2048)
  sock, err := net.ListenUDP("udp", &net.UDPAddr {
    IP:   net.ParseIP("127.0.0.1"),
    Port: client.port,
  })

  if err != nil {
    fmt.Printf("Listen error: %v\n", err)
    return
  }

  client.sock = sock
  defer sock.Close()

  for {
    n, err := sock.Read(buffer)
    if err != nil {
      fmt.Printf("Receive error: %v\n", err)
      continue
    }

    message := new(Message)
    err = proto.Unmarshal(buffer[:n], message)
    if err != nil {
      fmt.Printf("Parse error: %v\n", err)
      continue
    }

    client.work <- message
  }
}

func (client *Client) Run() {
  go client.Handle()
  go client.Listen()

  for {
    client.Send(&Message {
      Type:  Message_S_PROPOSE,
      Node:  uint32(client.port),
      Epoch: 0,
      Value: 0,
    }, client.port)

    time.Sleep(time.Second)
  }
}

func (client *Client) Send(message *Message, port int) {
  if client.sock == nil {
    fmt.Printf("Socket not yet open\n")
    return
  }

  buffer, err := proto.Marshal(message)
  if err != nil {
    fmt.Printf("Marshal error: %v\n", err)
    return
  }

  _, err = client.sock.WriteToUDP(buffer, &net.UDPAddr {
    IP:   net.ParseIP("127.0.0.1"),
    Port: port,
  })

  if err != nil {
    fmt.Printf("Send error: %v\n", err)
    return
  }
}

func NewClient(port int) Client {
  return Client {
    port: port,
    sock: nil,
    work: make(chan *Message, 16),
    logs: list.New(),
  }
}
