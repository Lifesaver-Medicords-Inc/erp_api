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

func GetInvTracker(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.InvTrackerView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item inventory tracker")
	}

	//Invalidate cache
	InvalidateItemCaches()

	return response, 0, nil
}

func GetInvWarehouseName(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.WarehouseNameView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item warehouse name")
	}

	return response, 0, nil
}

func CreateInvTracker(c *fiber.Ctx, tx *gorm.DB) (models.InvTracker, int, error) {
	var body models.InvTracker

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating inventory tracker")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	fmt.Println("at  ok ", at)

	if !ok {
		at = models.At{}
		fmt.Println("at not ok ", at)

	}

	atdata := models.InvTrackerAt{RefId: body.ID, InvTrackerContent: body.InvTrackerContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating inventorytrackerAt")
	}

	if err := services.InvalidateCacheByModel(models.InvTrackerView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, 0, nil
}

func UpdateInvTracker(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.InvTracker, int, error) {

	var body models.InvTracker

	if err := c.BodyParser(&body); err != nil {
		fmt.Println("This is the error:", err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Update record in DB
	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating Inventory Tracker")
	}

	//Save audit trail
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.InvTrackerAt{
		RefId:             body.ID,
		InvTrackerContent: body.InvTrackerContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating inventory trackerat")
	}

	if err := services.InvalidateCacheByModel(models.InvTrackerView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, 0, nil
}

func DeleteInvTracker(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.InvTracker, int, error) {
	var body models.InvTracker

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Printf("Parsed Body: %+v\n", body)

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting inventory tracker")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.InvTrackerAt{
		RefId:             body.ID,
		InvTrackerContent: body.InvTrackerContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating inventory tracker at")
	}

	InvalidateItemCaches()

	return body, 0, nil
}
