package topic_sockets

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

type manager interface {
	Topic() Topic
	HandleEvent(TopicEvent, chan<- TopicEvent)
}

type Hub struct {
	managers             map[Topic]manager
	connections          map[*connection]bool
	register, unregister chan *connection
	handleReceive        chan clientTopicEvent
}

func NewHub() *Hub {
	return &Hub{
		managers:      make(map[Topic]manager),
		connections:   make(map[*connection]bool),
		register:      make(chan *connection),
		unregister:    make(chan *connection),
		handleReceive: make(chan clientTopicEvent),
	}
}

func (h *Hub) AddManager(m manager) {
	h.managers[m.Topic()] = m
}

func (h *Hub) Listen(addr string) {
	go h.startBroker()

	r := mux.NewRouter()
	r.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		h.register <- NewClient(w, r, h.handleReceive, h.unregister)
	})
	err := http.ListenAndServe(addr, handlers.CombinedLoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (h *Hub) startBroker() {
	for {
		select {
		case conn := <-h.register:
			h.connections[conn] = true
		case conn := <-h.unregister:
			delete(h.connections, conn)
		case event := <-h.handleReceive:
			h.dispatchEvent(event)
		}
	}
}

func (h *Hub) dispatchEvent(event clientTopicEvent) {
	m := h.managers[event.Topic]
	m.HandleEvent(event.TopicEvent, event.Client.Sink)
}
