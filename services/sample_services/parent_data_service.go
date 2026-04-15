package sample_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Parent
	Childf models.Childf `json:"childf"`
	Childs models.Childs `json:"childs"`
}

func GetParents(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Parents []models.Parent `json:"parents"`
		Childfs []models.Childf `json:"childfs"`
		Childss []models.Childs `json:"childss"`
	}

	var response Response

	if err := services.DbGet(&response.Parents, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting parents")
	}

	if err := GetChildfs(&response.Childfs, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetChildss(&response.Childss, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetParent(id int) (Body, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var record Body

	if err := services.DbGet(&record.Parent, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting parent")
	}

	conditions = map[string]interface{}{
		"parent_id": record.ID,
	}

	if err := GetChildf(&record.Childf, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := GetChilds(&record.Childs, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

func CreateParent(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.Parent); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parent")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.ParentAt{RefId: body.ID, ParentContent: models.ParentContent{Name: body.Name, Description: body.Description}, At: at}
	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	if err := CreateChildf(tx, body.ID, body.Childf, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateChilds(tx, body.ID, body.Childs, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateParent(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.Parent, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating parent")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ParentAt{RefId: body.ID, ParentContent: models.ParentContent{Name: body.Name, Description: body.Description}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	conditions = map[string]interface{}{
		"parent_id": body.ID,
	}

	if err := UpdateChildf(tx, body.Childf, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateChilds(tx, body.Childs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func DeleteParent(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.Parent, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting parent")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ParentAt{RefId: body.ID, ParentContent: body.ParentContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	conditions = map[string]interface{}{
		"parent_id": body.ID,
	}

	if err := DeleteChildf(tx, body.Childf, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteChilds(tx, body.Childs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
