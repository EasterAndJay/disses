package main

import(
  "fmt"
  "log"
  "math/rand"
  "net"
  "os"
  "time"
)

const PEERS_FILE = "peers.txt"
const BASE_PORT = 5000

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
  rand.Seed(int64(os.Getpid()))
  clusterSize, id := parseArgs()
  port  := uint32(id + BASE_PORT)
  peers := make(map[uint32]*net.UDPAddr)

  ips, err := readLines(PEERS_FILE)
  if err != nil {
    log.Fatal("Error reading configuration file")
  }

  for i, ip := range ips[:clusterSize] {
    peerIP := net.ParseIP(ip)
    peerPort := i+BASE_PORT
    fmt.Printf("Adding ip addr to peers: %s:%d\n", ip, peerPort)
    peers[uint32(i)] = &net.UDPAddr {
      IP: peerIP,
      Port: peerPort,
    }
  }

  if id == -1 {
    runAll(peers)
  } else {
    runOne(peers, port)
  }
}
