package lamport

import(
  "github.com/xavierholt/disses/lamport-mutex/go/message"
)


type Queue []Message

func (q Queue) Len() int {
  return len(q)
}

func (q Queue) Swap(i, j int) {
  q[i], q[j] = q[j], q[i]
}

func (q Queue) Less(i, j int) bool {
  if q[i].Clock == q[j].Clock {
    return q[i].Pid < q[j].Pid
  }
  return q[i].Clock < q[j].Clock
}
