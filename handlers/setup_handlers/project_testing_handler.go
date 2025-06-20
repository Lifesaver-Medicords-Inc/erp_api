package setup_handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetProjects(c *fiber.Ctx) error {
	data, status, err := setup_services.GetProjects(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateProject(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.CreateProject(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	BroadcastProjects()

	return utils.RespondSuccess(c, data)
}

func UpdateProject(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateProject(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	BroadcastProjects()

	return utils.RespondSuccess(c, data)
}

func BroadcastProjects() error {
	data, status, err := setup_services.GetProjects(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastMessage(data); err != nil {
		return err
	}

	return nil
}

func WsgetProjects(c *websocket.Conn) {
	data, status, err := setup_services.GetProjects(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := services.BroadcastMessage(data); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Status:", status)

	BroadcastProjects()
}

func WsGetIdk(c *websocket.Conn, branch string, ProjectId string) {

	fmt.Println("WSGETIDK TO", branch)
	fmt.Println("WSGETIDK TO", ProjectId)

	data, status, err := setup_services.GetProjects(nil)
	if err != nil {
		fmt.Println("Error getting initial data:", err)
		return
	}

	if err := services.BroadcastToProject(branch, ProjectId, data); err != nil {
		fmt.Println("Error broadcasting initial data:", err)
		return
	}

	fmt.Println("Status:", status)

	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Println("Received raw message:", string(msgBytes))

		var message interface{}

		if err := json.Unmarshal(msgBytes, &message); err != nil {
			fmt.Println("Invalid JSON received:", err)
			continue
		}

		fmt.Println("Decoded message:", message)

		// data, status, err := setup_services.GetProjects(nil)
		// if err != nil {
		// 	fmt.Println("Error getting updated data:", err)
		// 	continue
		// }

		// if err := services.BroadcastToProject(branch, ProjectId, message); err != nil {
		// 	fmt.Println("Error broadcasting updated data:", err)
		// 	continue
		// }

		fmt.Println("Updated status:", status)
	}
}
