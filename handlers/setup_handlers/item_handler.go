package setup_handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetItems(c *fiber.Ctx) error {
	data, status, err := setup_services.GetItems(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := setup_services.GetItem(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := setup_services.CreateItem(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	go broadcastItems()

	return utils.RespondSuccess(c, data)
}

func UpdateItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateItem(c, tx, nil)

	fmt.Println("Update Body:", data)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func DeleteItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.DeleteItem(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func broadcastItems() error {
	data, status, err := setup_services.GetItems(nil)
	if err != nil {
		fmt.Println(status, err)
		return err
	}

	items, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling users:", err)
		return err
	}

	initializers.WM.RLock()
	defer initializers.WM.RUnlock()

	for client := range initializers.WM.Clients {
		if err := client.WriteMessage(websocket.TextMessage, items); err != nil {
			log.Println("error sending message:", err)
		}
	}

	return nil
}

func WsgetItems(c *websocket.Conn) {
	initializers.WM.AddClient(c)

	fmt.Println("Client Connected:", c.IP())

	broadcastItems()

	// Read messages from the client
	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		fmt.Println("Message Type:", msgType)
		fmt.Println("Raw Message:", string(msg))

		// Print the received message
		fmt.Printf("Received message: %s\n", msg)

		nmsg := fmt.Sprintf("Changed Message %s", msg)

		// Send the message back to the client
		if err := c.WriteMessage(msgType, []byte(nmsg)); err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}

	initializers.WM.RemoveClient(c)
	// Connection closed
	fmt.Println("Client disconnected:", c.IP())
}
