package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

 
func GetUserPermission(conditions map[string]interface{}) (adminmodels.UserPermission, int, error) {

	var permissions adminmodels.UserPermission

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, fiber.StatusInternalServerError, errors.New("failed getting user permission")
	}

	return permissions, 0, nil
}

func CreateUserPermission(c *fiber.Ctx, tx *gorm.DB) (adminmodels.UserPermission, int, error) {

	var body adminmodels.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating user permission")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{
		RefId: body.ID,
		UserPermission: adminmodels.UserPermission{
			UserId:    body.UserId,
			CanCreate: body.CanCreate,
			CanUpdate: body.CanUpdate,
			CanDelete: body.CanDelete,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating user permissionnat")
	}

	return body, 0, nil
}

func UpdateUserPermission(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.UserPermission, int, error) {

	var body adminmodels.UserPermission

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating user permission")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{
		RefId: body.ID,
		UserPermission: adminmodels.UserPermission{
			UserId:    body.UserId,
			CanCreate: body.CanCreate,
			CanUpdate: body.CanUpdate,
			CanDelete: body.CanDelete,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating user permissionat")
	}

	return body, 0, nil
}

func DeleteUserPermission(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.UserPermission, int, error) {

	var body adminmodels.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting user permission")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{
		RefId: body.ID,
		UserPermission: adminmodels.UserPermission{
			UserId:    body.UserId,
			CanCreate: body.CanCreate,
			CanUpdate: body.CanUpdate,
			CanDelete: body.CanDelete,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating user permissionat")
	}

	return body, 0, nil
}
