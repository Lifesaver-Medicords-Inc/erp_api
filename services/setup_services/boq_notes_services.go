package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BoqNoteBody struct {
	models.BoqNotes `json:"boq_notes"`
	// BoqNotes []models.BoqNotes `json:"boq_notes"`
}

func GetBoqNotes(conditions map[string]interface{}) ([]models.BoqNotes, int, error) {
	var boqnote []models.BoqNotes

	if err := services.DbGet(&boqnote, conditions); err != nil {
		return boqnote, fiber.StatusInternalServerError, errors.New("failed getting boq note")
	}

	return boqnote, 0, nil
}

func GetBoqNote(id int) (models.BoqNotes, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var boqnote models.BoqNotes

	if err := services.DbGet(&boqnote, conditions); err != nil {
		return boqnote, fiber.StatusInternalServerError, errors.New("failed getting boq note")
	}

	return boqnote, 0, nil
}

func CreateBoqNote(tx *gorm.DB, boqnote *models.BoqNotes) (int, error) {
	if err := services.DbInsert(tx, boqnote); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating boq note")
		}

		return fiber.StatusInternalServerError, err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	at = models.At{}
	// }

	// atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: at}
	// if err := services.DbInsert(tx, &atdata); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	// }

	return 0, nil
}

func UpdateBoqNote(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BoqNoteBody, int, error) {
	var body BoqNoteBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.BoqNotes, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating BoqNotes")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.BoqNotesAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating BoqNotesAt")
	}

	return body, 0, nil
}
func DeleteBoqNote(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.BoqNotes, int, error) {
	var body models.BoqNotes
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

	atdata := models.BoqNotesAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating SetupItemBomAt")
	}

	return body, 0, nil
}
