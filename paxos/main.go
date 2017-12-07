package main

import(
  "log"
  "net"
  "strings"
  "strconv"
)

const PEERS_FILE = "peers.txt"

func main() {
  ipPorts, err := readLines(PEERS_FILE)
  if err != nil {
    log.Fatal("Error reading configuration file")
  }
  port, err := strconv.Atoi(ipPorts[0])
  if err != nil {
    log.Fatal("Error invalid port on first line of configuration file")
  }
  peers := make(map[int32]*net.UDPAddr)
  for _, ipPort := range ipPorts[1:] {
    ip := net.ParseIP(strings.Split(ipPort, ":")[0])
    port, err := strconv.Atoi(strings.Split(ipPort, ":")[1])
    if err != nil {
      log.Fatal("Error invalid port in configuration file")
    }
    peers[int32(port)] = &net.UDPAddr {
      IP: ip,
      Port: port,
    }
  }
  client := NewClient(int32(port), peers)
  client.Run()
}

