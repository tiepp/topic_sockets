package topic_sockets

import (
	"net/http"
)

type connection struct {
	Sink       chan TopicEvent
	transport  *transport
	unregister chan<- *connection
}

func NewClient(w http.ResponseWriter, r *http.Request, onReceive chan<- clientTopicEvent, unregister chan<- *connection) *connection {
	conn := &connection{
		Sink:       make(chan TopicEvent, 256),
		transport:  NewTransport(w, r),
		unregister: unregister,
	}

	go conn.read(onReceive)
	go conn.write()

	return conn
}

func (c *connection) read(onReceive chan<- clientTopicEvent) {
	defer c.close()

	for {
		event, ok := c.transport.receiveEvent()
		if ok {
			onReceive <- clientTopicEvent{
				Client:     c,
				TopicEvent: event,
			}
		}
	}
}

func (c *connection) write() {
	defer c.close()

	for {
		event, ok := <-c.Sink
		if !ok {
			c.transport.close()
			return
		}
		go c.transport.sendEvent(event)
	}
}

func (c *connection) close() {
	close(c.Sink)
	c.transport.close()
	c.unregister <- c
}
