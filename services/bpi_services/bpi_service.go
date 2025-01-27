package bpi_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Bpi
	IndustriesId []int `json:"industries_id"`
}

func CreateBpi(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {

	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("body", body.IndustriesId)

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
		fmt.Println("INDUSTRIES", v)
		if err := CreateBpiIndustries(tx, body.ID, uint(v), at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil

}
