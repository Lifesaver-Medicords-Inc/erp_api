package setup_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetWarehouseNames(conditions map[string]interface{}) ([]models.WarehouseName, int, error) {
	var warehouse_name []models.WarehouseName

	if err := services.DbGet(&warehouse_name, conditions); err != nil {
		return warehouse_name, fiber.StatusInternalServerError, errors.New("failed getting warehouse names")
	}

	return warehouse_name, 0, nil
}

func GetWarehouseName(id int) (models.WarehouseName, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var warehouse_name models.WarehouseName

	if err := services.DbGet(&warehouse_name, conditions); err != nil {
		return warehouse_name, fiber.StatusInternalServerError, errors.New("failed getting warehouse name")
	}

	return warehouse_name, 0, nil
}

func CreateWarehouseName(c *fiber.Ctx, tx *gorm.DB) (models.WarehouseName, int, error) {
	var body models.WarehouseName
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating warehouse name")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WarehouseNameAt{RefId: body.ID, Code: body.Code, WarehouseNameContent: body.WarehouseNameContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehousenameat")
	}

	return body, 0, nil
}

func UpdateWarehouseName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseName, int, error) {
	var body models.WarehouseName
	// fmt.Println("body: ", body)

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("conditions: ", conditions)
	fmt.Println("&body: ", &body)

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		fmt.Println("body: ", body)
		return body, fiber.StatusInternalServerError, errors.New("failed updating warehousename")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WarehouseNameAt{
		RefId:                body.ID,
		Code:                 body.Code,
		WarehouseNameContent: body.WarehouseNameContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusBadRequest, errors.New("failed creating warehousenameat")
	}
	return body, 0, nil
}

func DeleteWarehouseName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseName, int, error) {
	var body models.WarehouseName
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting warehousename")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WarehouseNameAt{RefId: body.ID, Code: body.Code, WarehouseNameContent: body.WarehouseNameContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehouseat")
	}

	return body, 0, nil
}

func GetWarehouseManagers(conditions map[string]interface{}) ([]models.User, int, error) {
	var users []models.User

	//back end filtering
	// Add condition for filtering by position containing "admin" or has "manager" text in their position (since idk what position will handle warehouse)
	// if _, exists := conditions["position"]; !exists {
	// 	conditions["position"] = map[string]interface{}{
	// 		"$like": "%admin%",
	// 	}
	// }

	if err := services.DbGet(&users, conditions); err != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed to get users") //warehouse manager
	}

	return users, 0, nil
}
