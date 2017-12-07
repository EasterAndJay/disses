package main

import "fmt"
import "github.com/golang/protobuf/proto"
import "math/rand"
import "net"
import "time"

type Client struct {
  // Connection Stuff
  port      int32;
  sock      *net.UDPConn;
  work      chan *Message;
  wish      chan int32;

  // Paxos Stuff
  // clientVal int32;
  // clientRID int32;
  ballotNum int32;
  acceptNum int32;
  acceptVal int32;
  // acceptRID int32;
  peers     map[int32]*net.UDPAddr;
  okays     map[int32]bool;
  logs      []int32;

  // Soccer Stuff
  remainder int32;
}

func NewClient(port int32, peers map[int32]*net.UDPAddr) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),

    // clientVal: 0,
    // clientRID: 0,
    ballotNum: 0,
    acceptNum: 0,
    acceptVal: 0,
    // acceptRID: 0,
    peers:     peers,
    okays:     map[int32]bool{},
    logs:      make([]int32, 0),

    remainder: 100,
  }
}

func (client *Client) Broadcast(message *Message) {
  for _, addr := range client.peers {
    client.Send(addr, message)
  }
}

func (client *Client) Commit() {
  if client.acceptVal < -5000 {
    // Remove a server that's leaving.
    delete(client.peers, -client.acceptVal)
  } else if client.acceptVal > 5000 {
    // Add a new server.
    client.peers[client.acceptVal] = parseAddr(PEERS_FILE, int(client.acceptVal))
  } else {
    // Adjust the remaining tickets.
    client.remainder += client.acceptVal
  }

  client.logs = append(client.logs, client.acceptVal)
  //TODO: If we added a node, tell it!

  // if client.acceptRID == client.clientRID {
  //   // Got our pet value committed!
  //   select {
  //   case next, ok := <-wish:
  //     client.clientRID = rand.Int31()
  //     client.clientVal = next
  //   default:
  //     client.clientRID = 0
  //     client.clientVal = 0
  //   }
  // }

  client.ballotNum = 0
  client.acceptNum = 0
  client.acceptVal = 0
  // client.acceptRID = 0

  for index, entry := range client.logs {
    fmt.Printf(" - %4d: %d\n", index, entry)
  }
}

func (client *Client) GetEpoch() int32 {
  return int32(len(client.logs))
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
        client.Send(client.peers[message.GetNode()], &Message {
          Type:  Message_NOTIFY,
          Epoch: message.GetEpoch(),
          Value: client.logs[message.GetEpoch()],
        })
      }
    } else if epoch > client.GetEpoch() {
      // We're out of date!
      // Dummy proposition to force an update.
      client.Send(client.peers[message.GetNode()], &Message {
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
    for _, peerAddr := range client.peers {
      time.Sleep(time.Duration(5 * rand.Float32()) * time.Second)
      client.Send(peerAddr, &Message {
        Type:  Message_PETITION,
        Epoch: client.GetEpoch(),
        Value: 10,
      })
    }
  }
}

func (client *Client) Send(addr *net.UDPAddr, message *Message) {
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

  _, err = client.sock.WriteToUDP(buffer, addr)

  if err != nil {
    fmt.Printf("Send error: %v\n", err)
    return
  }
}
