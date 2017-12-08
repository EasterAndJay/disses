package main

import "fmt"
import "math/rand"
import "time"

func (client *Client) HandlePETITION(message* Message) {
  value := message.GetValue()
  if client.values[*value] {
    // Already have this one.
    return
  }

  go func() {
    time.Sleep(time.Duration(rand.Float32() * 5) * time.Second)
    client.Send(client.GetID(), message)
  }()

  client.acceptNum += 1
  client.Broadcast(&Message {
    Type:   Message_ACCEPT,
    Epoch:  client.GetEpoch(),
    Sender: client.GetID(),
    Ballot: client.acceptNum,
    Value:  value,
  })
}

func (client *Client) HandleACCEPT(message* Message) {
  if message.GetBallot() >= client.acceptNum {
    client.acceptNum = message.GetBallot()
    reply := client.MakeReply(Message_ACCEPTED, message)
    client.Broadcast(reply)
  } else {
    // client.Log("Ignoring ACCEPT: %v", message)
  }
}

func (client *Client) HandleACCEPTED(message* Message) {
  value := *message.GetValue()
  okays := client.accepts[value]
  if okays == nil {
    okays = make(map[uint32]bool)
    client.accepts[value] = okays
  }

  okays[message.GetSender()] = true
  if len(okays) > len(client.peers) / 2 {
    client.Commit(message)
  }
}

func (client *Client) HandleNOTIFY(message* Message) {
  client.Commit(message)
}

func (client *Client) HandleQUERY(message* Message) {
  reply := client.MakeReply(Message_NOTIFY, message)
  reply.Value = client.logs[message.GetEpoch()]
  client.Send(message.GetSender(), reply)
}

func (client *Client) HandleLOG(message* Message) {
  logstr := "The log as I see it:\n"
  for index, entry := range client.logs {
    logstr += fmt.Sprintf("%4d: %v\n", index, entry)
  }
  client.Log(logstr)
}
