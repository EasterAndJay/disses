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

  // Paxos Stuff
  acceptNum uint32;
  clientVal *Value;
  clientSeq uint32;

  peers     map[uint32]*net.UDPAddr;
  accepts   map[Value]map[uint32]bool;
  values    map[Value]bool;
  logs      []*Value;

  // Soccer Stuff
  tickets   uint32;
}

func NewClient(port uint32, peers map[uint32]*net.UDPAddr) Client {
  return Client {
    port:      port,
    sock:      nil,
    work:      make(chan *Message, 16),

    acceptNum: 0,
    clientVal: nil,
    clientSeq: 0,

    peers:     peers,
    accepts:   make(map[Value]map[uint32]bool),
    values:    make(map[Value]bool),
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
  client.values[*entry] = true

  switch(entry.GetType()) {
  case Value_BUY:
    if value <= client.tickets {
      client.tickets -= value
      if entry.GetClient() == client.GetID() {
        client.Log("Sold %d tickets: %d remaining.", value, client.tickets)
      }
    } else if entry.GetClient() == client.GetID() {
      client.Log("Can't sell %d tickets: only %d available.", value, client.tickets)
    }
  case Value_SUPPLY:
    client.tickets += value
    client.Log("Got %d new tickets: now have %d.", value, client.tickets)
  case Value_JOIN:
    client.peers[value] = parseAddr(PEERS_FILE, int(value))
  case Value_LEAVE:
    delete(client.peers, value)
  }

  // Clear out the values from the old epoch:
  client.accepts   = make(map[Value]map[uint32]bool)
  client.acceptNum = 0
  client.clientVal = nil
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
      case Message_LOG:
        client.HandleLOG(message)
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
      case Message_ACCEPT:
        client.HandleACCEPT(message)
      case Message_ACCEPTED:
        client.HandleACCEPTED(message)
      case Message_NOTIFY:
        client.HandleNOTIFY(message)
      case Message_QUERY:
        client.HandleQUERY(message)
      case Message_LOG:
        client.HandleLOG(message)
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

func (client *Client) Run() {
  go client.Handle()
  go client.Listen()

  for {
    time.Sleep(time.Duration(5 * rand.Float32()) * time.Second)
    client.Send(client.GetID(), &Message {
      Type:  Message_PETITION,
      Epoch: client.GetEpoch(),
      Value: &Value {
        Type:     Value_BUY,
        Client:   client.GetID(),
        Sequence: client.Sequence(),
        Value:    uint32(rand.Int31n(9) + 1),
      },
    })
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
