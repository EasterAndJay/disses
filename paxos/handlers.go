package main

func (client *Client) HandlePETITION(message* Message) {
  // INCOMPLETE
  value := message.GetValue()
  client.wishlist = append(client.wishlist, value)

  client.Broadcast(&Message {
    Type:   Message_PROPOSE,
    Epoch:  client.GetEpoch(),
    Ballot: client.ballotNum + 1,
    Value:  message.GetValue(),
  })
}

func (client *Client) HandlePROPOSE(message *Message) {
  if message.GetBallot() > client.ballotNum {
    client.ballotNum = message.GetBallot()

    reply := client.MakeReply(Message_PROMISE, message)
    reply.Accept = client.acceptNum
    reply.Value  = client.acceptVal
    client.Send(message.GetSender(), reply)
  } else {
    client.Log("Ignoring PROPOSE: %v", message)
  }
}

func (client *Client) HandlePROMISE(message* Message) {
  ballotNum := message.GetBallot()
  acceptNum := message.GetAccept()
  if _, ok := client.promises[ballotNum]; !ok {
    client.promises[ballotNum] = make(map[uint32]bool)
  }
  client.promises[ballotNum][message.GetSender()] = true
  if acceptNum > client.clientNum {
    client.clientVal = message.GetValue()
    client.clientNum = acceptNum
  }
  if len(client.promises[ballotNum]) > len(client.peers) / 2 {
    //TODO Yay acceptance
    reply := client.MakeReply(Message_ACCEPT, message)
    reply.Value = client.clientVal
    client.Broadcast(reply)
  }
}

func (client *Client) HandleACCEPT(message* Message) {
  if message.GetBallot() >= client.ballotNum {
    client.acceptNum = message.GetBallot()
    client.acceptVal = message.GetValue()

    reply := client.MakeReply(Message_ACCEPTED, message)
    client.Send(message.GetSender(), reply)
  } else {
    client.Log("Ignoring ACCEPT: %v", message)
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
    reply := client.MakeReply(Message_NOTIFY, message)
    client.Broadcast(reply)
  }
}

func (client *Client) HandleNOTIFY(message* Message) {
  client.acceptNum = message.GetBallot()
  client.acceptVal = message.GetValue()
  client.Commit(message)
}

func (client *Client) HandleQUERY(message* Message) {
  reply := client.MakeReply(Message_NOTIFY, message)
  reply.Value = client.logs[message.GetEpoch()]
  client.Send(message.GetSender(), reply)
}
