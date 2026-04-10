package purchasing_handlers

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/purchasing_services"
	"github.com/pierceperado/smpc/utils"
)

func GetPurchasingRedboxList(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetSortedPurchasingRedboxList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func BroadcastRedboxList() error {
	data, status, err := purchasing_services.GetSortedPurchasingRedboxList(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastMessage(data); err != nil {
		return err
	}

	return nil
}

func WsgetRedboxList(c *websocket.Conn) {
	// Send initial data on connect
	data, _, err := purchasing_services.GetSortedPurchasingRedboxList(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := services.SendMessage(c, data); err != nil {
		fmt.Println(err)
		return
	}

	// Keep connection alive, handle ping
	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", err)
			return
		}

		if msgType == websocket.TextMessage && string(msg) == "ping" {
			if err := c.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				fmt.Println("Pong failed:", err)
				return
			}
		}
	}
}
