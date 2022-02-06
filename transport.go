package topic_sockets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO actually check origin
		return true
	},
}

type transport struct {
	conn       *websocket.Conn
	pingTicker *time.Ticker
}

func NewTransport(w http.ResponseWriter, r *http.Request) *transport {
	conn, ok := createSocket(w, r)
	if !ok {
		return nil
	}
	return &transport{conn: conn, pingTicker: time.NewTicker(pingPeriod)}
}

func createSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	conn.SetReadLimit(maxMessageSize)
	err = conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return nil, false
	}

	conn.SetPongHandler(func(string) error {
		err = conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})

	return conn, true
}

func (c *transport) receiveEvent() (TopicEvent, bool) {
	event := TopicEvent{}
	err := c.conn.ReadJSON(&event)
	if err != nil {
		log.Printf("error: %v", err)
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("error: %v", err)
		}
		return event, false
	}
	return event, true
}

func (c *transport) sendEvent(event TopicEvent) bool {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		fmt.Println(err)
		return false
	}
	if err = c.conn.WriteJSON(event); err != nil {
		fmt.Println(err)
	}
	return false
}

func (c *transport) close() {
	err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		log.Println(err)
	}
}

func (c *transport) sendPing() bool {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		log.Println(err)
		return false
	}
	if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return true
	}
	return false
}
