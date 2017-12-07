package main

import(
  "flag"
  "log"
  "net"
  "strings"
  "strconv"
)

const PEERS_FILE = "peers.txt"

func main() {
  var clusterSize int
  flag.IntVar(&clusterSize, "cluster-size", 3, "Enter the cluster size. Valid values are one of: {3, 5}")
  if clusterSize != 3 || clusterSize != 5 {
    log.Fatal("Invalid cluster size")
  }
  ipPorts, err := readLines(PEERS_FILE)
  if err != nil {
    log.Fatal("Error reading configuration file")
  }
  port, err := strconv.Atoi(ipPorts[0])
  if err != nil {
    log.Fatal("Error invalid port on first line of configuration file")
  }
  peers := make(map[int32]*net.UDPAddr)
  for _, ipPort := range ipPorts[1:clusterSize+1] {
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

