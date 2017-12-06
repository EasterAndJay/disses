package main

import "fmt"
import "github.com/golang/protobuf/proto"
import "net"
import "time"

type Client struct {
  // Connection Stuff
  port      int32;
  sock      *net.UDPConn;
  work      chan *Message;

  // Paxos Stuff
  ballotNum int32;
  acceptNum int32;
  acceptVal int32;
  peers     map[int32]bool;
  okays     map[int32]bool;
  logs      []int32;

  // Soccer Stuff
  remainder int32;
}

func NewClient(port int32) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),

    ballotNum: 0,
    acceptNum: 0,
    acceptVal: 0,
    peers:     make(map[int32]bool),
    okays:     make(map[int32]bool),
    logs:      make([]int32, 0),

    remainder: 100,
  }
}

func (client *Client) Broadcast(message *Message) {
  for peer, _ := range client.peers {
    client.Send(peer, message)
  }
}

func (client *Client) Commit() {
  if client.acceptVal < -5000 {
    // Remove a server that's leaving.
    delete(client.peers, -client.acceptVal)
  } else if client.acceptVal > 5000 {
    // Add a new server.
    client.peers[client.acceptVal] = true
  } else {
    // Adjust the remaining tickets.
    client.remainder += client.acceptVal
  }

  client.logs = append(client.logs, client.acceptVal)
  client.ballotNum = 0
  client.acceptNum = 0
  client.acceptVal = 0
}

func (client *Client) GetEpoch() int32 {
  return int32(len(client.peers))
}

func (client *Client) GetID() int32 {
  return client.port
}

func (client *Client) Handle() {
  for {
    message := <-client.work
    epoch := message.GetEpoch()

    fmt.Printf("Got a message! {%v}\n", message)

    if epoch < client.GetEpoch() {
      // It's old. Reply with a NOTIFY.
      if message.GetType() != Message_NOTIFY {
        client.Send(message.GetNode(), &Message {
          Type:  Message_NOTIFY,
          Epoch: message.GetEpoch(),
          Value: client.logs[message.GetEpoch()],
        })
      }
    } else if epoch > client.GetEpoch() {
      // We're out of date!
      // Dummy proposition to force an update.
      client.Send(message.GetNode(), &Message {
        Type:  Message_PROPOSE,
        Epoch: client.GetEpoch(),
      })
    } else {
      switch message.GetType() {
      case Message_PETITION:
        client.HandlePETITION(message)
      case Message_PROPOSE:
        client.HandlePROPOSE(message)
      case Message_PROMISE:
        client.HandlePROMISE(message)
      case Message_ACCEPT:
        client.HandleACCEPT(message)
      case Message_ACCEPTED:
        client.HandleACCEPTED(message)
      case Message_NOTIFY:
        client.HandleNOTIFY(message)
      default:
        fmt.Printf("Unknown message type: %v\n", message)
      }
    }
  }
}

func (client *Client) Listen() {
  buffer := make([]byte, 2048)
  sock, err := net.ListenUDP("udp", &net.UDPAddr {
    IP:   net.ParseIP("127.0.0.1"),
    Port: int(client.port),
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
    time.Sleep(time.Second)
    client.Send(client.port, &Message {
      Type:  Message_PETITION,
      Value: 10,
    })
  }
}

func (client *Client) Send(port int32, message *Message) {
  if client.sock == nil {
    fmt.Printf("Socket not yet open\n")
    return
  }

  message.Node = client.GetID()
  buffer, err := proto.Marshal(message)
  if err != nil {
    fmt.Printf("Marshal error: %v\n", err)
    return
  }

  _, err = client.sock.WriteToUDP(buffer, &net.UDPAddr {
    IP:   net.ParseIP("127.0.0.1"),
    Port: int(port),
  })

  if err != nil {
    fmt.Printf("Send error: %v\n", err)
    return
  }
}
