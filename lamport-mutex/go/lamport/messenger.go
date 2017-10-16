package lamport

import(
  "encoding/gob"
  "fmt"
  "math"
  "net"
  "sort"
)

const (
  REQUEST = iota
  REPLY
  RELEASE
)

type Messenger struct {
  pid int
  *Connector
  clock int
  replyCount int

  queue Queue
  likeLock chan int
}

func (m *Messenger) SendMessage(msg Message, conn net.Conn) error {
  encoder := gob.NewEncoder(conn)
  m.UpdateClock(-1)
  err := encoder.Encode(msg)
  if err != nil {
    fmt.Println(err)
  }
  return err
}

func (m *Messenger) RecvMessage(conn net.Conn) (Message, error) {
  msg := Message{}
  decoder := gob.NewDecoder(conn)
  err := decoder.Decode(&msg)
  if err != nil {
    fmt.Println(err)
    return msg, err
  }
  m.UpdateClock(msg.Clock)
  return msg, nil
}

func (m *Messenger) Reply(conn net.Conn) {
  msg := Message{REPLY, m.pid, m.clock, -1}
  m.SendMessage(msg, conn)
}

func (m *Messenger) Request(conn net.Conn) {
  msg := Message{REQUEST, m.pid, m.clock, -1}
  m.Enqueue(msg)
  m.SendMessage(msg, conn)
}

func (m *Messenger) Release(conn net.Conn, likes int) {
  msg := Message{RELEASE, m.pid, m.clock, likes}
  m.queue = m.queue[1:]
  m.SendMessage(msg, conn)
}

func (m *Messenger) Enqueue(msg Message) {
  m.queue = append(m.queue, msg)
  sort.Sort(m.queue)
}

func (m *Messenger) UpdateClock(peerClock int) {
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
      if m.replyCount == len(m.connections) && m.queue[0].Pid == m.pid {
        m.replyCount = 0
        m.likeLock <- 1
      }
    case RELEASE:
      fmt.Printf("Client %d: Release Message received from Client %d\n", m.pid, senderPid)
      fmt.Printf("Old queue = %v, likes = %d\n", m.queue, *likes)
      *likes += 1
      m.queue = m.queue[1:]
      fmt.Printf("New queue = %v, likes = %d\n", m.queue, *likes)
      if m.replyCount == len(m.connections) && m.queue[0].Pid == m.pid {
        m.replyCount = 0
        m.likeLock <- 1
      }
    default:

    }
  }
}
