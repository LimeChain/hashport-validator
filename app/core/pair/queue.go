package pair

type Message struct {
	Payload interface{}
}

// Queue is a wrapper of a go channel, particularly to restrict actions on the channel itself
type Queue struct {
	channel chan *Message
}

// Push pushes a image to the channel
func (q *Queue) Push(message *Message) {
	q.channel <- message
}

func NewQueue() *Queue {
	ch := make(chan *Message)
	return &Queue{channel: ch}
}
