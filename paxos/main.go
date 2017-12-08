package main

import(
  "flag"
  "fmt"
  "github.com/golang/protobuf/proto"
  "log"
  "math/rand"
  "net"
  "os"
  "time"
)

const PEERS_FILE = "peers.txt"
const BASE_PORT = 5000

func send(peers map[uint32]*net.UDPAddr, id int, message *Message) {
  if id == -1 {
    id = 0
  }

  conn, err := net.DialUDP("udp", nil, peers[uint32(id)])
  if err != nil {
    fmt.Println("DialUDP error: %v\n", err)
    return
  }

  defer conn.Close()
  buffer, err := proto.Marshal(message)

  if err != nil {
    fmt.Printf("Marshal error: %v", err)
    return
  }

  _, err = conn.Write(buffer)
  if err != nil {
    fmt.Printf("Write error: %v", err)
    return
  }
}

func petition(peers map[uint32]*net.UDPAddr, id int, vtype Value_Type, value uint32) {
  send(peers, id, &Message {
    Type:  Message_PETITION,
    Value: &Value {
      Type:  vtype,
      Value: value,
    },
  })
}

func runOne(peers map[uint32]*net.UDPAddr, port uint32) {
  client := NewClient(port, peers)
  client.Run()
}

func runAll(peers map[uint32]*net.UDPAddr) {
  for id, _ := range peers {
    client := NewClient(id + BASE_PORT, peers)
    go client.Run()
  }

  for {
    time.Sleep(time.Second)
  }
}

func main() {
  n      := flag.Int("cluster-size", 3,     "Enter the cluster size. Valid values are one of: {3, 5}")
  id     := flag.Int("id",          -1,     "Enter the node id. Should be in range {0..4}")
  buy    := flag.Int("buy",          0,     "Number of tickets to buy.")
  supply := flag.Int("supply",       0,     "Number of tickets to add to the supply.")
  print  := flag.Bool("log",         false, "Print the log as one client sees it.")
  flag.Parse()

  if *n != 3 && *n != 5 {
    log.Fatal("Invalid cluster size: ", n)
  }

  ips, err := readLines(PEERS_FILE)
  if err != nil {
    log.Fatal("Error reading configuration file")
  }

  peers := make(map[uint32]*net.UDPAddr)
  for i, ip := range ips[:*n] {
    addr, err := net.ResolveUDPAddr("udp", ip)
    if err != nil {
      fmt.Printf("Resolve failure: %v\n", err)
      continue
    }

    peers[uint32(i)] = addr
  }

  if *buy > 0 {
    petition(peers, *id, Value_BUY, uint32(*buy))
  } else if *supply > 0 {
    petition(peers, *id, Value_SUPPLY, uint32(*supply))
  } else if *print {
    send(peers, *id, &Message {Type: Message_LOG})
  } else {
    if *id == -1 {
      runAll(peers)
    } else {
      rand.Seed(int64(os.Getpid()))
      runOne(peers, uint32(*id + BASE_PORT))
    }
  }
}
