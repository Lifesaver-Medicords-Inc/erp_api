package services

import (
	"errors"
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/pierceperado/smpc/initializers"
)

func HandleWs(c *websocket.Conn, h func(*websocket.Conn)) {
	initializers.WM.AddClient(c)
	fmt.Println("Client Connected:", c.IP())

	defer func() {
		initializers.WM.RemoveClient(c)
		fmt.Println("Client Disconnected:", c.IP())
	}()

	h(c)
}

func BroadcastMessage(data interface{}) error {
	initializers.WM.RLock()
	defer initializers.WM.RUnlock()

	for client := range initializers.WM.Clients {
		if err := client.WriteJSON(data); err != nil {
			return errors.New("error sending message")
		}
	}

	return nil
}
