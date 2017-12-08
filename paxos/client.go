package main

import "fmt"
import "github.com/golang/protobuf/proto"
import "math/rand"
import "net"
import "time"

const HEARTBEAT_INTERVAL = 500
const TIMEOUT = 5

type Client struct {
  // Connection Stuff
  port      uint32;
  sock      *net.UDPConn;
  work      chan *Message;
  wishlist  []*Value;

  // Paxos Stuff
  leaderID  int32;
  ballotNum uint32;
  acceptNum uint32;
  acceptVal *Value;

  clientVal *Value;
  clientNum uint32;

  clientSeq uint32;

  peers     map[uint32]*net.UDPAddr;
  promises  map[uint32]map[uint32]bool;
  accepts   map[Value]map[uint32]bool;
  logs      []*Value;
  heartbeat chan uint32;

  // Soccer Stuff
  tickets   uint32;
}

func NewClient(port uint32, peers map[uint32]*net.UDPAddr) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),
    wishlist:  make([]*Value, 0),

    leaderID:  -1,
    ballotNum: 0,
    acceptNum: 0,
    acceptVal: nil,
    clientVal: nil,
    clientNum: 0,

    peers:     peers,
    promises:  make(map[uint32]map[uint32]bool),
    accepts:   make(map[Value]map[uint32]bool),
    logs:      make([]*Value, 0),
    heartbeat: make(chan uint32, 1),
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
  // if len(client.logs) > 0 {
  //   lastEntry := client.logs[len(client.logs)-1]
  //   if lastEntry.Client == entry.Client && lastEntry.Sequence == entry.Sequence {
  //     // Already processed this message e.g. with new clients from a config change
  //     return
  //   }
  // }
  client.logs = append(client.logs, entry)

  switch(entry.GetType()) {
  case Value_BUY:
    if value < client.tickets {
      client.tickets -= value
    }
  case Value_SUPPLY:
    client.tickets += value
  case Value_JOIN:
    addr, err := parseAddr(PEERS_FILE, value)
    if err != nil {
      client.Log("Error parsing addr from config file: %v", err)
      break
    }
    client.peers[value] = addr
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
    client.Propose(client.wishlist[0])
  }

  // Clear out the values from the old epoch:
  client.promises  = make(map[uint32]map[uint32]bool)
  client.accepts   = make(map[Value]map[uint32]bool)
  client.ballotNum = 0
  client.acceptNum = 0
  client.acceptVal = nil
  client.clientVal = nil
  client.clientNum = 0

  logstr := "Committed:\n"
  // for index, entry := range client.logs {
  //   logstr += fmt.Sprintf(" - %4d: %v\n", index, entry)
  // }
  logstr += fmt.Sprintf(" - %4d: %v\n", len(client.logs)-1, entry)
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
    // client.Log("Got a message! {%v}", message)

    if message.GetEpoch() < client.GetEpoch() {
      switch message.GetType() {
      case Message_PETITION:
        client.HandlePETITION(message)
      case Message_NOTIFY:
        // Ignore it.
      default:
        client.Send(message.GetSender(), &Message {
          Type:  Message_NOTIFY,
          Epoch: message.GetEpoch(),
          Value: client.logs[message.GetEpoch()],
        })
      }
    } else if message.GetEpoch() > client.GetEpoch() {
      // We're out of date!
      // Dummy proposition to force an update.
      client.Send(message.GetSender(), &Message {
        Type:   Message_QUERY,
        Epoch:  client.GetEpoch(),
        Sender: client.GetID(),
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
      case Message_HEARTBEAT:
        client.HandleHEARTBEAT(message)
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

func (client *Client) Propose(value *Value) {
  client.ballotNum += 1
  client.Broadcast(&Message {
    Type:   Message_PETITION,
    Epoch:  client.GetEpoch(),
    Sender: client.GetID(),
    Ballot: client.ballotNum,
    Value:  value,
  })
}
func (client *Client) Join() {
  client.Propose(&Value {
    Type:     Value_JOIN,
    Client:   client.GetID(),
    Sequence: client.Sequence(),
    Value:    client.GetID(),
  })
}

func (client *Client) Run(joining bool) {
  go client.Handle()
  go client.Listen()
  if joining {
    client.leaderID = 0
    time.Sleep(time.Duration(2 * time.Second))
    client.Join()
  }
  for {
    time.Sleep(time.Duration(5 * rand.Float32()) * time.Second)
    // client.Send(client.GetID(), &Message {
    //   Type:  Message_PETITION,
    //   Epoch: client.GetEpoch(),
    //   Value: &Value {
    //     Type:     Value_BUY,
    //     Client:   client.GetID(),
    //     Sequence: client.Sequence(),
    //     Value:    uint32(rand.Int31n(10) + 1),
    //   },
    // })
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

func (client *Client) Heartbeat() {
  for {
    time.Sleep(HEARTBEAT_INTERVAL * time.Millisecond)
    if client.leaderID != int32(client.GetID()) {
      return
    }
    client.Broadcast(&Message {
      Type:   Message_HEARTBEAT,
      Epoch:  client.GetEpoch(),
      Sender: client.GetID(),
      Ballot: client.ballotNum,
      Value:  nil,
    })
  }
}

func (client *Client) StartTimeout(leader int32) {
  for {
    select {
    case <-client.heartbeat:
    case <-time.After(time.Second * TIMEOUT):
        if leader == client.leaderID {
          fmt.Println("===== TIMEOUT =====")
          client.leaderID = -1
        }
        return
    }
  }
}
