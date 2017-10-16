package lamport

import(
  "fmt"
  "os"
  "net"
  "time"
  "strconv"
  "sync"
)

const (
  CONN_HOST = "localhost"
  CONN_TYPE = "tcp"
  BASE_PORT = 5000
)

type Connector struct {
  connLock *sync.RWMutex
  connections map[int]net.Conn

  signal chan int
}


func (cn *Connector) AcceptConnections(pid int, n int) {
  port := strconv.Itoa(pid + BASE_PORT)
  l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+port)
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    os.Exit(1)
  }
  defer l.Close()
  fmt.Println("Listening on " + CONN_HOST + ":" + port)
  for {
    conn, err := l.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      os.Exit(1)
    }
    buff := make([]byte, 1024)
    read, err := conn.Read(buff)
    if err != nil {
      fmt.Println(err)
    }
    peerPid, err := strconv.Atoi(string(buff[:read]))
    if err != nil {
      fmt.Println(err)
    }
    if !cn.finishedConnecting(n) {
      cn.connLock.Lock()
      cn.addConnection(peerPid, conn)
      cn.connLock.Unlock()
    } else {
      break
    }
  }
}

func (cn *Connector) InitiateConnections(pid int, n int) {
  peerPid := -1
  for {
    peerPid += 1
    if peerPid > n - 1 {
      peerPid = 0
    }
    if peerPid == pid {
      continue
    }
    cn.connLock.Lock()
    if !cn.connectionExists(peerPid) {
      conn, err := cn.connectToPeer(peerPid)
      if err != nil {
        fmt.Printf("Client %d: Dial failed: %v\n", pid, err.Error())
        cn.connLock.Unlock()
        time.Sleep(time.Second)
        continue
      }
      sendPid(pid, conn)
      cn.addConnection(peerPid, conn)
    }
    cn.connLock.Unlock()
    if cn.finishedConnecting(n) {
      break
    }
  }
}

func (cn *Connector) connectionExists(peerPid int) bool {
  if _, ok := cn.connections[peerPid]; ok {
    return true
  }
  return false
}

func (cn *Connector) addConnection(peerPid int, conn net.Conn) {
  fmt.Printf("Adding peer with pid %d to connections\n", peerPid)
  cn.connections[peerPid] = conn
}

func (cn *Connector) connectToPeer(peerPid int) (net.Conn, error) {
  port := strconv.Itoa(BASE_PORT + peerPid)
  servAddr := CONN_HOST + ":" + port
  tcpAddr, err := net.ResolveTCPAddr(CONN_TYPE, servAddr)
  if err != nil {
    fmt.Println("ResolveTCPAddr failed:", err.Error())
    os.Exit(1)
  }
  return net.DialTCP(CONN_TYPE, nil, tcpAddr)
}

func (cn *Connector) finishedConnecting(n int) bool {
  cn.connLock.RLock()
  defer cn.connLock.RUnlock()
  if len(cn.connections) == n - 1 {
    cn.Signal()
    return true
  }
  return false
}

func (cn *Connector) Signal() {
  select {
  case cn.signal <- 1:
  default:
    close(cn.signal)
  }
}

func sendPid(pid int, conn net.Conn) {
  fmt.Printf("Client %d: Sending pid to peer\n", pid)
  _, err := writeInt(pid, conn)
  if err != nil {
    fmt.Println("Write to server failed:", err.Error())
    os.Exit(1)
  }
}

func writeInt(i int, conn net.Conn) (int, error) {
  return conn.Write([]byte(strconv.Itoa(i)))
}