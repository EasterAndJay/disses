type Entry struct {
 ballotNum int;
 acceptNum int;
 acceptVal *Message;
}

//TODO

func NewEntry(message* Message) {
  ...
}

func (entry *Entry) Handle(message *Message) {
  switch message.GetType() {
  case Message_PETITION(message):
    entry.HandlePETITION(message)
  case Message_PROPOSE:
    entry.HandlePROPOSE(message)
  case Message_ACCEPT:
    entry.HandleACCEPT(message)
  default:
    fmt.Printf("The hell is that?\n")
  }
}

func HandlePETITION(message* Message) {
  switch message.GetPhase() {
  case Message_REQUEST:
    fmt.Printf("Got a petition request\n")
  case Message_RESPONSE:
    fmt.Printf("Got a petition response\n")
  case Message_TIMEOUT:
    fmt.Printf("Got a petition timeout\n")
  default:
    fmt.Printf("The hell is that?\n")
  }
}

func HandlePROPOSE(message *Message) {
  switch message.GetPhase() {
  case Message_REQUEST:
    fmt.Printf("Got a propose request\n")
  case Message_RESPONSE:
    fmt.Printf("Got a propose response\n")
  case Message_TIMEOUT:
    fmt.Printf("Got a propose timeout\n")
  default:
    fmt.Printf("The hell is that?\n")
  }
}

func HandleACCEPT(message* Message) {
  switch message.GetPhase() {
  case Message_REQUEST:
    fmt.Printf("Got an accept request\n")
  case Message_RESPONSE:
    fmt.Printf("Got an accept response\n")
  case Message_TIMEOUT:
    fmt.Printf("Got an accept timeout\n")
  default:
    fmt.Printf("The hell is that?\n")
  }
}
