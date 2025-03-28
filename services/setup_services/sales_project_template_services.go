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
	fmt.Println("DATAAAA", body)
	for _, v := range body.SalesProjectTemplateChild {
		if err := CreateSalesProjectTemplateChild(tx, body.TemplateID, v, at); err != nil {
			fmt.Println("DATAAAA", body)
			fmt.Println("ERRRRRR", err)
			return body, fiber.StatusInternalServerError, errors.New("failed creating project template child")
		}
	}

	return body, 0, nil
}

func CreateSalesProjectTemplateChild(tx *gorm.DB, parentID uint, TemplateChild models.SalesProjectTemplateChild, at models.At) error {

	TemplateChild.BasedID = parentID

	if err := services.DbInsert(tx, &TemplateChild); err != nil {
		return errors.New("failed creating template child")
	}

	templatechildat := models.SalesProjectTemplateChildAt{
		RefId:                            TemplateChild.NodesID,
		SalesProjectTemplateChildContent: TemplateChild.SalesProjectTemplateChildContent,
		At:                               at,
	}

	if err := services.DbInsert(tx, &templatechildat); err != nil {
		return errors.New("failed creating templatechildat")
	}

	return nil
}
