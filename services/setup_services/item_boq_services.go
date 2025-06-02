package setup_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BoqBody struct {
	models.ItemBoq
	BoqDetails         []models.ItemBoqDetails `json:"boq_details"`
	ProjectComponent   []models.ProjectComponent
	SalesProjectWiring []models.SalesProjectWiring
}

func GetItemBoqs(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ProjectComponent   []models.ProjectComponent   `json:"project_component"`
		SalesProjectWiring []models.SalesProjectWiring `json:"sales_project_wiring"`
	}

	var response Response

	if err := services.DbGet(&response.ProjectComponent, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}
	if err := services.DbGet(&response.SalesProjectWiring, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}
	return response, 0, nil
}

func GetQQnotes(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		QQView []models.QQView `json:"qq_view"`
	}

	var response Response

	if err := services.DbGet(&response.QQView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}
	return response, 0, nil
}

func CreateItemBoq(c *fiber.Ctx, tx *gorm.DB) (models.ItemBoqDetails, int, error) {
	var body models.ItemBoqDetails

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating canvas sheet")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemBoqDetailsAt{RefId: body.ID, ItemBoqDetailsContent: body.ItemBoqDetailsContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating boq at")
	}
	key := services.GetKey(models.ProjectComponent{}, nil)
	services.InvalidateCache(key)
	key2 := services.GetKey(models.QQView{}, nil)
	services.InvalidateCache(key2)
	return body, 0, nil
}
func UpdateItemBoq(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ItemBoqDetails, int, error) {
	var body models.ItemBoqDetails
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("Body:", body)

	conditions = map[string]interface{}{
		"id": body.ID,
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating CRM")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemBoqDetailsAt{RefId: body.ID, ItemBoqDetailsContent: body.ItemBoqDetailsContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating CRMat")
	}
	fmt.Println("body: ", body)
	key := services.GetKey(models.ProjectComponent{}, nil)
	services.InvalidateCache(key)
	key2 := services.GetKey(models.QQView{}, nil)
	services.InvalidateCache(key2)
	return body, 0, nil
}
