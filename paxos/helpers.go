package main

import(
  "bufio"
  "flag"
  "log"
  "net"
  "os"
  "strconv"
  "strings"
)

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

func parseAddr(file string, port int) *net.UDPAddr {
  portStr := strconv.Itoa(port)
  f, err := os.Open(file)
  if err != nil {
    log.Fatal("Peers file does not exist")
    return nil
  }
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    text := scanner.Text()
    if strings.Contains(text, portStr) {
      addr, err := net.ResolveUDPAddr("udp", text)
      if err != nil {
        log.Fatal("Error parsing ip:port address from peers file")
        return nil
      }
      return addr
    }
  }
  log.Fatal("ip:port not found in peers file")
  return nil
}

func parseArgs() (int, int) {
  var clusterSize int
  var id int
  flag.IntVar(&clusterSize, "cluster-size", 3, "Enter the cluster size. Valid values are one of: {3, 5}")
  flag.IntVar(&id, "id", -1, "Enter the node id. Should be in range {0..4}")
  flag.Parse()
  if clusterSize != 3 && clusterSize != 5 {
    log.Fatal("Invalid cluster size: ", clusterSize)
  }
  if id < 0 || id > 4 {
    log.Fatal("Invalid id: ", id)
  }
   return clusterSize, id 
}
