package pair

type Watcher interface {
	Watch(queue *Queue)
}

type Handler interface {
	Handle(interface{})
}

type Pair struct {
	queue   *Queue
	watcher Watcher
	handler Handler
}

func (p *Pair) Listen() {
	p.handle()
	p.watch()
}

func (p *Pair) handle() {
	go func() {
		for messages := range p.queue.channel {
			p.handler.Handle(messages.Payload)
		}
	}()
}

func (p *Pair) watch() {
	p.watcher.Watch(p.queue)
}

func NewPair(watcher Watcher, handler Handler) *Pair {
	return &Pair{
		watcher: watcher,
		handler: handler,
		queue:   NewQueue(),
	}
}
