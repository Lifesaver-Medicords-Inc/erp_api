package hub

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	Send    chan []byte
	Channel string
}

func (c *Client) ReadPump(h *Hub) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var base BaseMessage
		if err := json.Unmarshal(data, &base); err != nil {
			log.Println("invalid message:", err)
			continue
		}

		switch base.Channel {
		case "chat":
			var chatMsg ChatMessage
			if err := json.Unmarshal(data, &chatMsg); err != nil {
				log.Println("invalid chat message:", err)
				continue
			}
			jsonMsg, _ := json.Marshal(chatMsg)
			h.Broadcast <- Message{Channel: "chat", Data: jsonMsg}

		case "orders":
			var orderMsg OrderMessage
			if err := json.Unmarshal(data, &orderMsg); err != nil {
				log.Println("invalid order message:", err)
				continue
			}
			jsonMsg, _ := json.Marshal(orderMsg)
			h.Broadcast <- Message{Channel: "orders", Data: jsonMsg}

		default:
			log.Println("unknown channel:", base.Channel)
		}
	}
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}
