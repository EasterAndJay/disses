package main

func main() {
  peers := map[uint32]bool{
    5000: true,
    5001: true,
    5002: true,
  }

  a := NewClient(5000, peers)
  b := NewClient(5001, peers)
  c := NewClient(5002, peers)

  go a.Run()
  go b.Run()
  c.Run()
}
