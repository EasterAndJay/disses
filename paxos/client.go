package main

import "fmt"
import "github.com/golang/protobuf/proto"
import "math/rand"
import "net"
import "time"

type Client struct {
  // Connection Stuff
  port      uint32;
  sock      *net.UDPConn;
  work      chan *Message;
  wish      chan int32;

  // Paxos Stuff
  ballotNum uint32;
  acceptNum uint32;
  acceptVal *Value;

  clientVal *Value;
  clientSeq uint32;

  peers     map[uint32]bool;
  promises  map[*Value]map[uint32]bool;
  accepts   map[*Value]map[uint32]bool;
  logs      []*Value;

  // Soccer Stuff
  remainder uint32;
}

func NewClient(port uint32) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),

    ballotNum: 0,
    acceptNum: 0,
    acceptVal: nil,
    clientVal: nil,

    peers:     map[uint32]bool{port: true},
    promises:  make(map[*Value]map[uint32]bool),
    accepts:   make(map[*Value]map[uint32]bool),
    logs:      make([]*Value, 0),

    remainder: 100,
  }
}

func (client *Client) Broadcast(message *Message) {
  for peer, _ := range client.peers {
    client.Send(peer, message)
  }
}

func (client *Client) Commit() {
  entry := client.acceptVal
  value := entry.GetValue()
  client.logs = append(client.logs, entry)

  switch(entry.GetType()) {
  case Value_BUY:
    if value < client.remainder {
      client.remainder -= value
    }
  case Value_SUPPLY:
    client.remainder += value
  case Value_JOIN:
    client.peers[value] = true
    //TODO: If we added a node, tell it!
  case Value_LEAVE:
    delete(client.peers, value)
  }

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

  // Clear out the values from the old epoch:
  client.promises  = make(map[*Value]map[uint32]bool)
  client.accepts   = make(map[*Value]map[uint32]bool)
  client.ballotNum = 0
  client.acceptNum = 0
  client.acceptVal = nil

  for index, entry := range client.logs {
    fmt.Printf(" - %4d: %d\n", index, entry)
  }
}

func (client *Client) GetEpoch() uint32 {
  return uint32(len(client.logs))
}

func (client *Client) GetID() uint32 {
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
        client.Send(message.GetSender(), &Message {
          Type:  Message_NOTIFY,
          Epoch: message.GetEpoch(),
          Value: client.logs[message.GetEpoch()],
        })
      }
    } else if epoch > client.GetEpoch() {
      // We're out of date!
      // Dummy proposition to force an update.
      client.Send(message.GetSender(), &Message {
        Type:  Message_QUERY,
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
      case Message_QUERY:
        client.HandleQUERY(message)
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

func (client *Client) MakeReply(mtype Message_Type, message* Message) *Message {
  return &Message {
    Type:   mtype,
    Epoch:  message.GetEpoch(),
    Sender: message.GetSender(),
    Ballot: message.GetBallot(),
    Value:  message.GetValue(),
  }
}

func (client *Client) Run() {
  go client.Handle()
  go client.Listen()

  for {
    time.Sleep(time.Duration(5 * rand.Float32()) * time.Second)
    client.Send(client.port, &Message {
      Type:  Message_PETITION,
      Epoch: client.GetEpoch(),
      Value: &Value {
        Type:     Value_BUY,
        Client:   client.GetID(),
        Sequence: client.Sequence(),
        Value:    uint32(rand.Int31n(10) + 1),
      },
    })
  }
}

func (client *Client) Send(port uint32, message *Message) {
  if client.sock == nil {
    fmt.Printf("Socket not yet open\n")
    return
  }

  message.Sender = client.GetID()
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

func (client *Client) Sequence() uint32{
  client.clientSeq += 1
  return client.clientSeq
}
