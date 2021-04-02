package pair

type Watcher interface {
	Watch(queue *Queue)
}

type Handler interface {
	Handle(interface{})
}

// Pair represents a pair of a watcher and a handler, to which the watcher pushes messages
// which the handler processes
type Pair struct {
	queue   *Queue
	watcher Watcher
	handler Handler
}

// Listen begins the actions of the handler and the watcher
func (p *Pair) Listen() {
	p.handle()
	p.watch()
}

// handle subscribes to the channel messages, processing them synchronously
func (p *Pair) handle() {
	go func() {
		for messages := range p.queue.channel {
			p.handler.Handle(messages.Payload)
		}
	}()
}

// watch initializes the Watcher's Watch
func (p *Pair) watch() {
	go p.watcher.Watch(p.queue)
}

func NewPair(watcher Watcher, handler Handler) *Pair {
	return &Pair{
		watcher: watcher,
		handler: handler,
		queue:   NewQueue(),
	}
}
