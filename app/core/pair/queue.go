package pair

type Message struct {
	Payload interface{}
}

type Queue struct {
	channel chan *Message
}

func (q *Queue) Push(message *Message) {
	q.channel <- message
}

func NewQueue() *Queue {
	ch := make(chan *Message)
	return &Queue{channel: ch}
}
