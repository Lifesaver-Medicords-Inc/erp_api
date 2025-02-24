package setup_handlers

import (
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
