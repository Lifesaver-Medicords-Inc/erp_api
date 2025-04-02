package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BoqBody struct {
	models.ItemBoq
	BoqDetails       []models.ItemBoqDetails `json:"boq_details"`
	ProjectComponent []models.ProjectComponent
}

func GetItemBoqs(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ProjectComponent []models.ProjectComponent `json:"project_component"`
	}

	var response Response

	if err := services.DbGet(&response.ProjectComponent, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}

	return response, 0, nil
}

func CreateItemBoq(c *fiber.Ctx, tx *gorm.DB) (BoqBody, int, error) {
	var body BoqBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.ProjectComponent); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating canvas sheet")
	}

	return body, 0, nil
}

func UpdateItemBoq(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BoqBody, int, error) {
	var body BoqBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.ItemBoq, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating setup item bom")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemBoqAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating SetupItemBomAt")
	}

	for _, v := range body.BoqDetails {

		if err := UpdateBoqChild(tx, v, at, body.ID); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

	}

	return body, 0, nil
}
