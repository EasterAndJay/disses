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

func petition(peers map[uint32]*net.UDPAddr, id int, vtype Value_Type, value uint32) {
  if id == -1 {
    id = 0
  }

  conn, err := net.DialUDP("udp", nil, peers[uint32(id)])
  if err != nil {
    fmt.Println("DialUDP error: %v\n", err)
    return
  }

  buffer, err := proto.Marshal(&Message {
    Type:  Message_PETITION,
    Value: &Value {
      Type:  vtype,
      Value: value,
    },
  })

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
  var clusterSize int
  var id          int
  var buy         int
  var supply      int

  flag.IntVar(&clusterSize, "cluster-size", 3, "Enter the cluster size. Valid values are one of: {3, 5}")
  flag.IntVar(&id,          "id",          -1, "Enter the node id. Should be in range {0..4}")
  flag.IntVar(&buy,         "buy",          0, "Number of tickets to buy.")
  flag.IntVar(&supply,      "supply",       0, "Number of tickets to add to the supply.")
  flag.Parse()

  if clusterSize != 3 && clusterSize != 5 {
    log.Fatal("Invalid cluster size: ", clusterSize)
  }

  ips, err := readLines(PEERS_FILE)
  if err != nil {
    log.Fatal("Error reading configuration file")
  }

  peers := make(map[uint32]*net.UDPAddr)
  for i, ip := range ips[:clusterSize] {
    addr, err := net.ResolveUDPAddr("udp", ip)
    if err != nil {
      fmt.Printf("Resolve failure: %v\n", err)
      continue
    }

    peers[uint32(i)] = addr
  }

  for id, addr := range peers {
    fmt.Printf("%v: %v\n", id, addr)
  }

  if buy > 0 {
    petition(peers, id, Value_BUY, uint32(buy))
  } else if supply > 0 {
    petition(peers, id, Value_SUPPLY, uint32(supply))
  } else {
    if id == -1 {
      runAll(peers)
    } else {
      rand.Seed(int64(os.Getpid()))
      runOne(peers, uint32(id + BASE_PORT))
    }
  }
}
