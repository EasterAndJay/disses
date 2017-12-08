package main

func (client *Client) HandlePETITION(message* Message) {
  // INCOMPLETE
  value := message.GetValue()
  client.wishlist = append(client.wishlist, value)
  switch client.leaderID {
  case int32(client.GetID()):
    // I am leader
    client.Broadcast(&Message {
      Type:   Message_ACCEPT,
      Epoch:  client.GetEpoch(),
      Ballot: client.ballotNum + 1,
      Value:  message.GetValue(),
    })
  case -1:
    client.Log("THERE ISN'T A LEADER - I WILL DO IT")
    // No leader
    client.Broadcast(&Message {
      Type:   Message_PROPOSE,
      Epoch:  client.GetEpoch(),
      Ballot: client.ballotNum + 1,
      Value:  message.GetValue(),
    })
  default:
    // Someone else is leader
    client.Send(uint32(client.leaderID), message)
  }
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
    if client.clientVal == nil {
      client.clientVal = client.wishlist[0]
    }

    reply := client.MakeReply(Message_ACCEPT, message)
    reply.Value = client.clientVal
    client.Broadcast(reply)
  }
}

func (client *Client) HandleACCEPT(message* Message) {
  if message.GetBallot() >= client.ballotNum {
    if client.leaderID != int32(message.GetSender()) {
      // New leader
      client.Log("New leader is client: %d", message.GetSender())
      if client.GetID() == message.GetSender() {
        // This client just became leader
        go client.Heartbeat()
      }
      client.leaderID = int32(message.GetSender())
      if len(client.heartbeat) > 0 {
        <-client.heartbeat
      }
      go client.StartTimeout(client.leaderID)
    }

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
  client.Commit(message)
}

func (client *Client) HandleQUERY(message* Message) {
  reply := client.MakeReply(Message_NOTIFY, message)
  reply.Value = client.logs[message.GetEpoch()]
  client.Send(message.GetSender(), reply)
}

func (client *Client) HandleHEARTBEAT(message* Message) {
  if(int32(message.GetSender()) == client.leaderID) {
    select {
    case client.heartbeat <- 1:
      client.Log("Put a heartbeat in the queue")
    default:
      client.Log("Heartbeat channel full - skipping this heartbeat message")
    }
  }
}
