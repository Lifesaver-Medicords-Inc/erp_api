package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetUsers(conditions map[string]interface{}, tx *gorm.DB) ([]models.User, int, error) {
	var users []models.User

	if err := tx.Where(conditions).Preload("Permissions").Preload("Position").Find(&users).Error; err != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed getting users")
	}

	for i := range users {
		users[i].Password = ""

	}

	return users, 0, nil
}

func GetUser(conditions map[string]interface{}, tx *gorm.DB) (models.User, int, error) {
	var user models.User

	if err := tx.Where(conditions).Preload("Permissions").Preload("Position").First(&user).Error; err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed getting users")
	}

	user.Password = ""

	return user, 0, nil
}

func UpdateUser(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.User, int, error) {

	var body models.User

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating user")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UserAt{RefId: body.ID, UserContent: models.UserContent{
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Password:   body.Password,
		PositionId: body.PositionId,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	return body, 0, nil
}

func DeleteUser(c *fiber.Ctx, tx *gorm.DB, Id int) (int, error) {

	if err := services.DbDelete(tx, &models.UserPermissionAt{}, map[string]interface{}{"ref_id": Id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user")
	}

	if err := services.DbDelete(tx, &models.UserPermission{}, map[string]interface{}{"user_id": Id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user")
	}

	if err := services.DbDelete(tx, &models.UserAt{}, map[string]interface{}{"ref_id": Id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user")
	}

	if err := services.DbDelete(tx, &models.User{}, map[string]interface{}{"id": Id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user")
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	at = models.At{}
	// }

	// atdata := models.UserAt{RefId: body.ID, At: at}
	// if err := services.DbInsert(tx, &atdata); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed creating userat")
	// }

	return 0, nil
}
