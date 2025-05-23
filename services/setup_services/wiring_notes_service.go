package setup_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetWiringNotes(conditions map[string]interface{}) (interface{}, int, error) {
	key := services.GetKey(models.WiringNotes{}, conditions)
	services.InvalidateCache(key)
	type Response struct {
		WritingNotes []models.WiringNotes `json:"wiring_notes"`
	}

	var response Response
	if err := services.DbGet(&response.WritingNotes, conditions); err != nil {
		fmt.Println("this is error: ", err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}
	return response, 0, nil
}

func CreateWiringNote(c *fiber.Ctx, tx *gorm.DB) (models.WiringUserInput, int, error) {
	var body models.WiringUserInput

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating wiring note")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WiringUserInputAt{RefId: body.NoteID, WiringUserInputContent: body.WiringUserInputContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating wiring note at")
	}
	fmt.Println("BODY IS: ", body)
	key := services.GetKey(models.ProjectComponent{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}
func UpdateWiringNote(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WiringUserInput, int, error) {
	var body models.WiringUserInput
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("Body:", body)

	conditions = map[string]interface{}{
		"note_id": body.NoteID,
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating CRM")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WiringUserInputAt{RefId: body.NoteID, WiringUserInputContent: body.WiringUserInputContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating wiring note at")
	}
	fmt.Println("BODY IS: ", body)
	key := services.GetKey(models.ProjectComponent{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}
