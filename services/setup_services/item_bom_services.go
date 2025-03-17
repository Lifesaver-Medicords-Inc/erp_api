package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BodyParse struct {
	models.SetupItemBom
	BomDetails []models.SetupItemBomDetails `json:"bom_details"`
}

func GetSetupItemBoms(conditions map[string]interface{}) (interface{}, int, error) {
	//var setupItemBoms []models.SetupItemBom

	type Response struct {
		ItemBom     []models.BomViewList     `json:"bom_head"`
		ItemDetails []models.BomViewItemList `json:"bom_details"`
	}
	var response Response

	if err := services.DbGet(&response.ItemBom, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bom")
	}

	if err := services.DbGet(&response.ItemDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bom")
	}

	return response, 0, nil
}

func GetBomItemList(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.BomView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi item list")
	}

	return response, 0, nil
}

func GetSetupItemBom(id int) (models.SetupItemBom, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var setupItemBom models.SetupItemBom

	if err := services.DbGet(&setupItemBom, conditions); err != nil {
		return setupItemBom, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return setupItemBom, 0, nil
}

func GetBomParentDetail(conditions map[string]interface{}) ([]models.BomViewList, int, error) {
	var BomViewList []models.BomViewList

	if err := services.DbGet(&BomViewList, conditions); err != nil {
		return BomViewList, fiber.StatusInternalServerError, errors.New("failed getting BomViewList")
	}
	return BomViewList, 0, nil
}

func GetBomChildDetail(conditions map[string]interface{}) ([]models.BomViewItemList, int, error) {
	var BomViewItemList []models.BomViewItemList

	if err := services.DbGet(&BomViewItemList, conditions); err != nil {
		return BomViewItemList, fiber.StatusInternalServerError, errors.New("failed getting BomViewItemList")
	}
	return BomViewItemList, 0, nil
}

func CreateSetupItemBom(c *fiber.Ctx, tx *gorm.DB) (BodyParse, int, error) {
	var body BodyParse
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SetupItemBom); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating bom")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SetupItemBomAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating SetupItemBomAt")
	}

	//Create Bom Details

	for _, v := range body.BomDetails {
		if err := BomDetails(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func UpdateSetupItemBom(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyParse, int, error) {
	var body BodyParse

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.SetupItemBom, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating setup item bom")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SetupItemBomAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating SetupItemBomAt")
	}

	for _, v := range body.BomDetails {

		if err := UpdateChild(tx, v, at, body.ID); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

	}

	return body, 0, nil
}

func DeleteSetupItemBom(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.SetupItemBom, int, error) {
	var body models.SetupItemBom
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting setup item bom")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SetupItemBomAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating SetupItemBomAt")
	}

	return body, 0, nil
}
