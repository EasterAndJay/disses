package lamport

import(
  "fmt"
  "math"
  "net"
  proto "github.com/golang/protobuf/proto"
  "sort"

  "github.com/xavierholt/disses/lamport-mutex/go/message"
)

const (
  ENQUEUE message.Message_Type = 0
  REQUEST message.Message_Type = 1
  REPLY   message.Message_Type = 2
  RELEASE message.Message_Type = 3
)



type Messenger struct {
  pid int
  *Connector
  clock int
  replyCount int

  queue message.Queue
  likeLock chan int
}

func (m *Messenger) SendMessage(msg message.Message, conn net.Conn) error {
  data, err := proto.Marshal(&msg)
  if err != nil {
    fmt.Println(err)
    return err
  }
  m.UpdateClock(0)
  if _, err := conn.Write(data); err != nil {
    fmt.Println(err)
  }
  return err
}

func (m *Messenger) RecvMessage(conn net.Conn) (message.Message, error) {
  msg := message.Message{}
  data := make([]byte, 1024)
  n, err := conn.Read(data)
  if err != nil {
    fmt.Println(err)
    return msg, err
  }
  if err := proto.Unmarshal(data[:n], &msg); err != nil {
    fmt.Println("Receive: ", err)
    return msg, err
  }
  m.UpdateClock(msg.Clock)
  return msg, nil
}

func (m *Messenger) Reply(conn net.Conn) {
  msg := message.Message{REPLY, uint32(m.pid), uint32(m.clock), 0}
  m.SendMessage(msg, conn)
}

func (m *Messenger) Request(conn net.Conn) {
  msg := message.Message{REQUEST, uint32(m.pid), uint32(m.clock), 0}
  m.Enqueue(msg)
  m.SendMessage(msg, conn)
}

func (m *Messenger) Release(conn net.Conn, likes int) {
  msg := message.Message{RELEASE, uint32(m.pid), uint32(m.clock), uint32(likes)}
  m.queue = m.queue[1:]
  m.SendMessage(msg, conn)
}

func (m *Messenger) Enqueue(msg message.Message) {
  m.queue = append(m.queue, msg)
  sort.Sort(m.queue)
}

func (m *Messenger) UpdateClock(peerClock uint32) {
  m.clock = int(math.Max(float64(m.clock), float64(peerClock))) + 1
  fmt.Printf("Client %d: Updated clock to %d\n", m.pid, m.clock)
}

func (m *Messenger) ProcessMsg(senderPid int, conn net.Conn, likes *int) {
  for {
    msg, err := m.RecvMessage(conn)
    if err != nil {
      panic(err)
    }
    switch msg.MsgType {
    case REQUEST:
      fmt.Printf("Client %d: Request Message received from Client %d\n", m.pid, senderPid)
      m.Enqueue(msg)
      m.Reply(conn)
    case REPLY:
      fmt.Printf("Client %d: Reply Message received from Client %d\n", m.pid, senderPid)
      m.replyCount += 1
      if m.replyCount == len(m.connections) && m.queue[0].Pid == uint32(m.pid) {
        m.replyCount = 0
        m.likeLock <- 1
      }
    case RELEASE:
      fmt.Printf("Client %d: Release Message received from Client %d\n", m.pid, senderPid)
      fmt.Printf("Old queue = %v, likes = %d\n", m.queue, *likes)
      *likes += 1
      m.queue = m.queue[1:]
      fmt.Printf("New queue = %v, likes = %d\n", m.queue, *likes)
      if m.replyCount == len(m.connections) && m.queue[0].Pid == uint32(m.pid) {
        m.replyCount = 0
        m.likeLock <- 1
      }
    default:

    }
  }
}
