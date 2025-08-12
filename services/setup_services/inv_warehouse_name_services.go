package setup_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

//trash old code, too laze to update

func GetWarehouseManagers(conditions map[string]interface{}) ([]models.User, int, error) {
	var users []models.User

	if err := services.DbGet(&users, conditions); err != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed to get users") //warehouse manager
	}

	return users, 0, nil
}

type WarehouseBody struct {
	WarehouseName    models.WarehouseName    `json:"warehouse_name"`
	WarehouseAddress models.WarehouseAddress `json:"warehouse_address"`
	WarehouseArea    models.WarehouseArea    `json:"warehouse_area"`
}

type GetSpecificWarehouseBody struct {
	WarehouseName    models.WarehouseName    `json:"warehouse_name"`
	WarehouseAddress models.WarehouseAddress `json:"warehouse_address"`
	WarehouseArea    []models.WarehouseArea  `json:"warehouse_area"`
}

func GetWarehouseNames(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		WarehouseName    []models.WarehouseName    `json:"warehouse_name"`
		WarehouseAddress []models.WarehouseAddress `json:"warehouse_address"`
		WarehouseArea    []models.WarehouseArea    `json:"warehouse_area"`
	}

	var response Response

	if err := services.DbGet(&response.WarehouseName, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting warehouse names")
	}

	if err := services.DbGet(&response.WarehouseAddress, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting warehouse address")
	}

	if err := services.DbGet(&response.WarehouseArea, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting warehouse area")
	}

	return response, 0, nil
}

func GetWarehouseName(id int) (GetSpecificWarehouseBody, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var body GetSpecificWarehouseBody

	if err := services.DbGet(&body.WarehouseName, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting warehouse name")
	}

	conditions = map[string]interface{}{
		"warehouse_name_id": body.WarehouseName.ID,
	}

	if err := GetWarehouseAddress(&body.WarehouseAddress, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	areas, _, err := GetWarehouseAreasRow(conditions) //get list of children (grid)
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	body.WarehouseArea = areas

	return body, 0, nil
}

func CreateWarehouseName(c *fiber.Ctx, tx *gorm.DB) (WarehouseBody, int, error) {
	var body WarehouseBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.WarehouseName); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehouse")
	}

	generatedWarehouseCode := utils.WarehouseCodeGenerator(body.WarehouseName.ID)
	body.WarehouseName.Code = generatedWarehouseCode

	// Only generate LocationCode if frontend didn't supply one (usually used by postman)
	if body.WarehouseArea.LocationCode == "" {
		body.WarehouseArea.LocationCode = utils.AreaCodeGenerator(
			body.WarehouseArea.Zone,
			body.WarehouseArea.Area,
			body.WarehouseArea.Rack,
			body.WarehouseArea.Level,
			body.WarehouseArea.Bins,
		)
	}

	if err := tx.Model(&body.WarehouseName).Update("code", body.WarehouseName.Code).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating warehouse name code")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	warehousenameat := models.WarehouseNameAt{
		RefId:                body.WarehouseName.ID,
		Code:                 body.WarehouseName.Code,
		WarehouseNameContent: body.WarehouseName.WarehouseNameContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &warehousenameat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehouse at")
	}

	body.WarehouseAddress.WarehouseNameId = body.WarehouseName.ID
	body.WarehouseArea.WarehouseNameId = body.WarehouseName.ID

	if err := CreateWarehouseAddress(tx, body.WarehouseName.ID, body.WarehouseAddress, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateWarehouseArea(tx, body.WarehouseName.ID, body.WarehouseArea, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateWarehouseName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (WarehouseBody, int, error) {
	var body WarehouseBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//front end na ung code generation (for backend testing purpose)
	generatedAreaCode := utils.AreaCodeGenerator(
		body.WarehouseArea.Zone,
		body.WarehouseArea.Area,
		body.WarehouseArea.Rack,
		body.WarehouseArea.Level,
		body.WarehouseArea.Bins)
	body.WarehouseArea.LocationCode = generatedAreaCode

	if err := services.DbUpdate(tx, &body.WarehouseName, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating warehouse name")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.WarehouseNameAt{
			RefId:                body.WarehouseName.ID,
			WarehouseNameContent: body.WarehouseName.WarehouseNameContent,
			At:                   at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehouse name at")
	}

	conditions = map[string]interface{}{
		"warehouse_name_id": body.WarehouseName.ID,
	}

	if err := UpdateWarehouseAddress(tx, body.WarehouseAddress, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func DeleteWarehouseName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (WarehouseBody, int, error) {
	var body WarehouseBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("DEL ERR", err)
		fmt.Println("DEL body", body)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.WarehouseName, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting parent")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.WarehouseNameAt{
			RefId:                body.WarehouseName.ID,
			WarehouseNameContent: body.WarehouseName.WarehouseNameContent,
			At:                   at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	conditions = map[string]interface{}{
		"warehouse_name_id": body.WarehouseName.ID,
	}

	if err := DeleteWarehouseAddress(tx, body.WarehouseAddress, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
