package adminhandlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

func CreateWarehouse(c *fiber.Ctx) error {
	var body models.WarehouseName
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := adminservices.CreateWarehouse(body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetWarehouse(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	position, status, err := adminservices.GetWarehouse(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, position)
}

func GetWarehouses(c *fiber.Ctx) error {
	id := c.Query("id")
	warehouseManager := c.Query("warehouse-manager")
	isInActiveStr := c.Query("is-inactive")
	code := c.Query("code")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)

	isInactive, err := strconv.ParseBool(isInActiveStr)

	if err == nil {
		conditions["is_inactive"] = isInactive
	}

	if idNum != 0 {
		conditions["id"] = id
	}

	if warehouseManager != "" {
		conditions["warehouse_manager"] = warehouseManager
	}

	if code != "" {
		conditions["code"] = code
	}

	vehicles, status, err := adminservices.GetWarehouses(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, vehicles)
}
