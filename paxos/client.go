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
  wishlist  []*Value;

  // Paxos Stuff
  ballotNum uint32;
  acceptNum uint32;
  acceptVal *Value;

  clientVal *Value;
  clientSeq uint32;

  peers     map[uint32]*net.UDPAddr;
  promises  map[Value]map[uint32]bool;
  accepts   map[Value]map[uint32]bool;
  logs      []*Value;

  // Soccer Stuff
  tickets   uint32;
}

func NewClient(port uint32, peers map[uint32]*net.UDPAddr) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),
    wishlist:  make([]*Value, 0),

    ballotNum: 0,
    acceptNum: 0,
    acceptVal: nil,
    clientVal: nil,

    peers:     peers,
    promises:  make(map[Value]map[uint32]bool),
    accepts:   make(map[Value]map[uint32]bool),
    logs:      make([]*Value, 0),

    tickets:   100,
  }
}

func (client *Client) Broadcast(message *Message) {
  for id, _ := range client.peers {
    client.Send(id, message)
  }
}

func (client *Client) Commit(message *Message) {
  entry := message.GetValue()
  value := entry.GetValue()
  client.logs = append(client.logs, entry)

  switch(entry.GetType()) {
  case Value_BUY:
    if value < client.tickets {
      client.tickets -= value
    }
  case Value_SUPPLY:
    client.tickets += value
  case Value_JOIN:
    client.peers[value] = parseAddr(PEERS_FILE, int(value))
  case Value_LEAVE:
    delete(client.peers, value)
  }

  // Remove wishes that have been committed:
  for len(client.wishlist) > 0 {
    if client.TrimWishlist() {
      break
    }
  }

  if len(client.wishlist) > 0 {
    go func {
      time.Sleep(time.Duration(2 * rand.Float32()) * time.Second)
      client.Propose(client.wishlist[0])
    }
  }

  // Clear out the values from the old epoch:
  client.promises  = make(map[Value]map[uint32]bool)
  client.accepts   = make(map[Value]map[uint32]bool)
  client.ballotNum = 0
  client.acceptNum = 0
  client.acceptVal = nil

  logstr := "Committed:\n"
  for index, entry := range client.logs {
    logstr += fmt.Sprintf(" - %4d: %v\n", index, entry)
  }
  client.Log(logstr)
}

func (client *Client) GetEpoch() uint32 {
  return uint32(len(client.logs))
}

func (client *Client) GetID() uint32 {
  return client.port - BASE_PORT
}

func (client *Client) Handle() {
  for {
    message := <-client.work
    epoch := message.GetEpoch()

    client.Log("Got a message! {%v}", message)

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
        client.Log("Unknown message type: %v", message)
      }
    }
  }
}

func (client *Client) Listen() {
  buffer := make([]byte, 2048)
  sock, err := net.ListenUDP("udp", &net.UDPAddr {
    IP:   net.ParseIP("0.0.0.0"),
    Port: int(client.port),
  })

  if err != nil {
    client.Log("Listen error: %v", err)
    return
  }

  client.sock = sock
  defer sock.Close()

  for {
    n, err := sock.Read(buffer)
    if err != nil {
      client.Log("Receive error: %v", err)
      continue
    }

    message := new(Message)
    err = proto.Unmarshal(buffer[:n], message)
    if err != nil {
      client.Log("Parse error: %v", err)
      continue
    }

    client.work <- message
  }
}

func (client *Client) Log(format string, args ...interface{}) {
  fmt.Printf("%d:  %s\n",  client.port, fmt.Sprintf(format, args...))
}

func (client *Client) MakeReply(mtype Message_Type, message* Message) *Message {
  return &Message {
    Type:   mtype,
    Epoch:  message.GetEpoch(),
    Sender: client.GetID(),
    Ballot: message.GetBallot(),
    Value:  message.GetValue(),
  }
}

func (client *Client) Propose(value Value) {
  client.ballotNum += 1
  client.Broadcast(&Message {
    Type:   Message_PROPOSE,
    Epoch:  client.GetEpoch(),
    Sender: client.GetID(),
    Ballot: client.ballotNum,
    Value:  message.GetValue(),
  })
}

func (client *Client) Run() {
  go client.Handle()
  go client.Listen()

  for {
    for peerid, _ := range client.peers {
      time.Sleep(time.Duration(5 * rand.Float32()) * time.Second)
      client.Send(peerid, &Message {
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
}

func (client *Client) Send(peerid uint32, message *Message) {
  if client.sock == nil {
    client.Log("Socket not yet open")
    return
  }

  addr := client.peers[peerid]
  message.Sender = client.GetID()
  buffer, err := proto.Marshal(message)
  if err != nil {
    client.Log("Marshal error: %v", err)
    return
  }

  _, err = client.sock.WriteToUDP(buffer, addr)

  if err != nil {
    client.Log("Send error: %v", err)
    return
  }
}

func (client *Client) Sequence() uint32{
  client.clientSeq += 1
  return client.clientSeq
}

func (client *Client) TrimWishlist() bool {
  for _, entry := range client.logs {
    if *entry == *client.wishlist[0] {
      client.wishlist = client.wishlist[1:]
      return false
    }
  }

  return true
}
