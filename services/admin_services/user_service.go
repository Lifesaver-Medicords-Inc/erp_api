package adminservices

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetUsers(conditions map[string]interface{}) ([]models.User, int, error) {
	var users []models.User

	if err := services.DbGetNoCache(&users, conditions); err != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed getting users")
	}

	for i := range users {
		users[i].Password = ""

	}

	return users, 0, nil
}

func GetUser(conditions map[string]interface{}) (models.User, int, error) {
	var user models.User

	if err := services.DbGetNoCache(&user, conditions); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed getting users")
	}

	user.Password = ""

	return user, 0, nil
}

func CreateUser(c *fiber.Ctx, tx *gorm.DB) (models.User, int, error) {
	var body models.User

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating user")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UserAt{RefId: body.ID, UserContent: models.UserContent{
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Position:   body.Position,
		Password:   body.Password,
		PositionId: body.PositionId,
	}, At: at}

	UserPermission := adminmodels.UserPermission{
		UserId:    body.ID,
		CanCreate: false,
		CanUpdate: false,
		CanDelete: false,
	}

	jsonBody, _ := json.Marshal(UserPermission)

	c.Request().SetBody(jsonBody)

	CreateUserPermission(c, tx)

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
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
		Position:   body.Position,
		Password:   body.Password,
		PositionId: body.PositionId,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	return body, 0, nil
}
