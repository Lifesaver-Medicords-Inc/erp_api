package bpi_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Bpi
	IndustriesId []uint                  `json:"industries_id"`
	General      models.BpiGeneralSchema `json:"general"`
	Contacts     models.BpiContacts      `json:"contacts"`
}

type BpiResponse struct {
	models.Bpi
	IndustryIds   string `json:"industry_ids"`
	IndustryNames string `json:"industry_names"`
}

func GetBpis(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Bpi []BpiResponse `json:"bpi"`
		// Childfs []models.Childf `json:"childfs"`
		// Childss []models.Childs `json:"childss"`
	}

	var response Response

	if err := services.DbRaw(&response.Bpi, "GetBpiList", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpis")
	}

	return response, 0, nil
}

func CreateBpi(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.Bpi); err != nil {
		return body, fiber.StatusInternalServerError, errors.New(("failed creating parent"))
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.BpiAt{RefId: body.ID, BpiContent: models.BpiContent{SalesId: body.SalesId, Name: body.Name, Tin: body.Tin, Tel_no: body.Tel_no}, At: at}
	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	for _, v := range body.IndustriesId {
		if err := CreateBpiIndustries(tx, body.ID, uint(v), at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	if err := BpiGeneral(tx, body.ID, body.General, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	if err := BpiContact(tx, body.ID, body.Contacts, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
