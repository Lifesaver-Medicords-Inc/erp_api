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

type BpiResponse struct {
	models.Bpi
	IndustryIds   []uint   `json:"industry_ids"`
	BpiId         uint     `json:"bpi_id"`
	IndustryNames []string `json:"industry_names"`
}

func GetBpis(conditions map[string]interface{}) ([]BpiResponse, int, error) {
	var bpisResponse []BpiResponse

	type BpiResult struct {
		models.Bpi
		IndustryId   uint   `json:"industry_id"`
		BpiId        uint   `json:"bpi_id"`
		IndustryName string `json:"industry_name"`
	}

	var bpiResults []BpiResult
	if err := services.DbRaw(&bpiResults, "GetBPIInfos", conditions); err != nil {
		return bpisResponse, fiber.StatusInternalServerError, errors.New("failed getting bpis")
	}

	groupedBpis := make(map[uint]*BpiResponse)

	for _, bpi := range bpiResults {
		if existingBpi, ok := groupedBpis[bpi.ID]; ok {
			existingBpi.IndustryIds = append(existingBpi.IndustryIds, bpi.IndustryId)
			existingBpi.IndustryNames = append(existingBpi.IndustryNames, bpi.IndustryName)
		} else {
			groupedBpis[bpi.ID] = &BpiResponse{
				Bpi:           bpi.Bpi,
				IndustryIds:   []uint{bpi.IndustryId},
				BpiId:         bpi.BpiId,
				IndustryNames: []string{bpi.IndustryName},
			}
		}
	}

	for _, bpi := range groupedBpis {
		bpisResponse = append(bpisResponse, *bpi)
	}

	return bpisResponse, 0, nil
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
