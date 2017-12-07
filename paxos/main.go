package main

import(
  "fmt"
  "log"
  "net"
)

const PEERS_FILE = "peers.txt"
const BASE_PORT = 5000

func main() {
  clusterSize, id := parseArgs()
  port := id + BASE_PORT
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
  client := NewClient(uint32(port), peers)
  client.Run()
}
