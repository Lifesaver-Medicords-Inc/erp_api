package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetSocial(conditions map[string]interface{}) ([]models.Social, int, error) {
	var social []models.Social

	if err := services.DbGet(&social, conditions); err != nil {
		return social, fiber.StatusInternalServerError, errors.New("failed getting social media")
	}

	return social, 0, nil
}

func GetSocialMedia(id int) (models.Social, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var social models.Social

	if err := services.DbGet(&social, conditions); err != nil {
		return social, fiber.StatusInternalServerError, errors.New("failed getting social media")
	}

	return social, 0, nil
}

func CreateSocial(c *fiber.Ctx, tx *gorm.DB) (models.Social, int, error) {
	var body models.Social
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating social media")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SocialAt{RefId: body.ID, Code: body.Code, SocialContent: models.SocialContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating social media at")
	}

	return body, 0, nil
}

func UpdateSocial(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Social, int, error) {
	var body models.Social
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating social media")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SocialAt{RefId: body.ID, Code: body.Code, SocialContent: models.SocialContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating social media")
	}

	return body, 0, nil
}

func DeleteSocial(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Social, int, error) {
	var body models.Social
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting social media")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SocialAt{RefId: body.ID, Code: body.Code, SocialContent: models.SocialContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating social media at")
	}

	return body, 0, nil
}
