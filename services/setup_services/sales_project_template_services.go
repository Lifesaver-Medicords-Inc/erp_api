package setup_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type CreateTemplateBody struct {
	models.SalesProjectTemplate
	SalesProjectTemplateChild []models.SalesProjectTemplateChild `json:"sales_project_template_child"`
}

func GetSalesProjectTemplates(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		SalesProjectTemplate      []models.SalesProjectTemplate
		SalesProjectTemplateChild []models.SalesProjectTemplateChild `json:"sales_project_template_child"`
	}

	var response Response

	if err := services.DbGet(&response.SalesProjectTemplate, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales project template")
	}

	if err := services.DbGet(&response.SalesProjectTemplateChild, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales project template child")
	}

	return response, 0, nil
}

func CreateSalesProjectTemplate(c *fiber.Ctx, tx *gorm.DB) (CreateTemplateBody, int, error) {
	var body CreateTemplateBody

	fmt.Println("RAW REQUEST BODY:", string(c.Body()))

	if err := c.BodyParser(&body); err != nil {
		fmt.Println("BODY PARSING ERROR:", err)
		fmt.Println("PARSED BODY:", body) // Check what was actually parsed
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("DATAAAA", body)
	if err := services.DbInsert(tx, &body.SalesProjectTemplate); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating projects template")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := models.SalesProjectTemplateAt{RefId: body.TemplateID, SalesProjectTemplateContent: body.SalesProjectTemplateContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	for _, v := range body.SalesProjectTemplateChild {
		if err := CreateSalesProjectTemplateChild(tx, body.TemplateID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed creating project template child")
		}
	}

	return body, 0, nil
}

func CreateSalesProjectTemplateChild(tx *gorm.DB, parentID uint, TemplateChild models.SalesProjectTemplateChild, at models.At) error {
	TemplateChild.ParentID = parentID

	if err := services.DbInsert(tx, &TemplateChild); err != nil {
		return errors.New("failed creating template child")
	}

	templatechildat := models.SalesProjectTemplateChildAt{
		RefId:                            TemplateChild.ParentID,
		SalesProjectTemplateChildContent: TemplateChild.SalesProjectTemplateChildContent,
		At:                               at,
	}

	if err := services.DbInsert(tx, &templatechildat); err != nil {
		return errors.New("failed creating templatechildat")
	}

	return nil
}

func UpdateSalesProjectTemplate(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	var body CreateTemplateBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if body.TemplateID == 0 {
		return body, fiber.StatusBadRequest, errors.New("missing template id")
	}

	// update main template fields
	if err := tx.Model(&models.SalesProjectTemplate{}).
		Where("template_id = ?", body.TemplateID).
		Updates(body.SalesProjectTemplate).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating projects template")
	}

	// create an "at" record for the update
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := models.SalesProjectTemplateAt{
		RefId:                       body.TemplateID,
		SalesProjectTemplateContent: body.SalesProjectTemplateContent,
		At:                          at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales project template at")
	}

	// replace children: delete existing children and their at records, then recreate from payload
	if err := tx.Where("parent_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplateChild{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting existing template children")
	}
	if err := tx.Where("ref_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplateChildAt{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting existing template child at records")
	}

	for _, v := range body.SalesProjectTemplateChild {
		if err := CreateSalesProjectTemplateChild(tx, body.TemplateID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed creating project template child")
		}
	}

	return body, 0, nil
}

func DeleteProjectTemplate(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	type Body struct {
		TemplateID uint `json:"template_id"`
	}
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if body.TemplateID == 0 {
		return body, fiber.StatusBadRequest, errors.New("missing template id")
	}

	// delete child "at" records first
	if err := tx.Where("ref_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplateChildAt{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting template child at records")
	}

	// delete children
	if err := tx.Where("parent_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplateChild{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting template children")
	}

	// delete template "at" records
	if err := tx.Where("ref_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplateAt{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting template at records")
	}

	// delete main template
	if err := tx.Where("template_id = ?", body.TemplateID).Delete(&models.SalesProjectTemplate{}).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting template")
	}

	return body, 0, nil
}
