package main

import "fmt"

func (client *Client) HandlePETITION(message* Message) {
  if message.GetValue() < 0 && message.GetValue() > -5000 {
    if client.remainder + message.GetValue() < 0 {
      // TODO: Reject!
    }
  }

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
    client.Send(message.GetNode(), &Message {
      Type:   Message_PROMISE,
      Epoch:  message.GetEpoch(),
      Ballot: message.GetBallot(),
      Value:  message.GetValue(),
    })
  } else {
    fmt.Printf("Ignoring PROPOSE: %v\n", message)
  }
}

func (client *Client) HandlePROMISE(message* Message) {
  // if client.state != PROPOSING {
  //   return
  // }

  client.okays[message.GetNode()] = true
  if len(client.okays) > len(client.peers) / 2 {
    //TODO Yay acceptance
    client.Send(message.GetNode(), &Message {
      Type:   Message_ACCEPT,
      Epoch:  message.GetEpoch(),
      Ballot: message.GetBallot(),
      Value:  message.GetValue(),
    })
  }
}

func (client *Client) HandleACCEPT(message* Message) {
  if message.GetBallot() >= client.ballotNum {
    client.acceptNum = message.GetBallot()
    client.acceptVal = message.GetValue()

    client.Send(message.GetNode(), &Message {
      Type:   Message_ACCEPTED,
      Epoch:  message.GetEpoch(),
      Ballot: message.GetBallot(),
      Value:  message.GetValue(),
    })
  } else {
    fmt.Printf("Ignoring ACCEPT: %v\n", message)
  }
}

func (client *Client) HandleACCEPTED(message* Message) {
  // if client.state != ACCEPTING {
  //   return
  // }

  client.okays[message.GetNode()] = true
  if len(client.okays) > len(client.peers) / 2 {
    client.Commit()

    client.Broadcast(&Message {
      Type:   Message_NOTIFY,
      Epoch:  message.GetEpoch(),
      Ballot: message.GetBallot(),
      Value:  message.GetValue(),
    })
  }
}

func (client *Client) HandleNOTIFY(message* Message) {
  client.acceptNum = message.GetBallot()
  client.acceptVal = message.GetValue()
  client.Commit()
}
