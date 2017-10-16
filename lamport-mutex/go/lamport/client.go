package lamport

import (
  "fmt"
  "net"
  "sync"
  "time"
)

type Client struct {
  likes int
  post string
  *Messenger
}

func NewClient(pid int, post string, n int) *Client {
  return &Client{
    0,
    post,
    &Messenger{
      pid,
      &Connector{
        &sync.RWMutex{},
        make(map[int]net.Conn),

        make(chan int, 1),
      },
      0,
      0,
      make(Queue, 0, n),
      make(chan int, 1),
    },
  }
}

func (c *Client) Run(n int) {
  fmt.Printf("Client %d: Running\n", c.pid)
  go c.AcceptConnections(c.pid, n)
  go c.InitiateConnections(c.pid, n)
  <- c.signal
  fmt.Printf("Client %d: Connected to all other peers\n", c.pid)
  go c.RecvMsgs()
  for {
    // fmt.Printf("Client %d: Post content -  %s | LIKE count: %d\n", c.pid, c.post, c.likes)
    time.Sleep(time.Second * 5)
    c.Like()
  }
}

func (c *Client) RecvMsgs() {
  for senderPid, conn := range c.connections {
    fmt.Printf("Client %d: Starting to process messages from client %d\n", c.pid, senderPid)
    go c.ProcessMsg(senderPid, conn, &c.likes)
  }
}

func (c *Client) Like() {
  c.RequestLock()
  c.likes += 1
  fmt.Printf("Client %d: LIKES = %d\n", c.pid, c.likes)
  c.ReleaseLock()
}

func (c *Client) RequestLock() {
  for senderPid, conn := range c.connections {
    fmt.Printf("Client %d: Sending request message to client %d\n", c.pid, senderPid)
    c.Request(conn)
  }
  <-c.likeLock
}

func (c *Client) ReleaseLock() {
  for senderPid, conn := range c.connections {
    fmt.Printf("Client %d: Sending release message to client %d\n", c.pid, senderPid)
    c.Release(conn, c.likes)
  }
}
