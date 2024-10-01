package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	
	writeWait = 10 * time.Second

	
	pongWait = 60 * time.Second

	
	pingPeriod = (pongWait * 9) / 10

	
	maxMessageSize = 512
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	Username string `json:"username"`
}

type Message struct {
	UnreadCount string
	Username    string
}

func (c *Client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		
		if !ok {
			return
		}

		c.Conn.WriteJSON(message)
	}
}

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}
