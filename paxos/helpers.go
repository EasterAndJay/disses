package main

import(
  "bufio"
  "net"
  "os"
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

func parseAddr(file string, id uint32) (*net.UDPAddr, error) {
  addrs, err := readLines(file)
  if err != nil {
    return nil, err
  }
  addr, err := net.ResolveUDPAddr("udp", addrs[id])
  if err != nil {
    return nil, err
  }

  return addr, nil
}
