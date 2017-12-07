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
    client.Send(message.GetSender(), reply)
  } else {
    client.Log("Ignoring PROPOSE: %v", message)
  }
}

func (client *Client) HandlePROMISE(message* Message) {
  value := *message.GetValue()
  okays := client.promises[value]
  if okays == nil {
    okays = make(map[uint32]bool)
    client.promises[value] = okays
  }

  okays[message.GetSender()] = true
  if len(okays) > len(client.peers) / 2 {
    //TODO Yay acceptance
    reply := client.MakeReply(Message_ACCEPT, message)
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
  okays := client.promises[value]
  if okays == nil {
    okays = make(map[uint32]bool)
    client.promises[value] = okays
  }

  okays[message.GetSender()] = true
  if len(okays) > len(client.peers) / 2 {
    client.Commit()
    reply := client.MakeReply(Message_NOTIFY, message)
    client.Broadcast(reply)
  }
}

func (client *Client) HandleNOTIFY(message* Message) {
  client.acceptNum = message.GetBallot()
  client.acceptVal = message.GetValue()
  client.Commit()
}

func (client *Client) HandleQUERY(message* Message) {
  reply := client.MakeReply(Message_NOTIFY, message)
  reply.Value = client.logs[message.GetEpoch()]
  client.Send(message.GetSender(), reply)
}
