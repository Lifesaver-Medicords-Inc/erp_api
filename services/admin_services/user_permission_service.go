package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPermission(conditions map[string]interface{}) (models.UserPermission, int, error) {

	var permissions models.UserPermission

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, fiber.StatusInternalServerError, errors.New("failed getting user permission")
	}

	return permissions, 0, nil
}

func GetPermissions(conditions map[string]interface{}) ([]models.UserPermission, int, error) {

	var permissions []models.UserPermission

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, fiber.StatusInternalServerError, errors.New("failed getting user permission")
	}

	return permissions, 0, nil
}

func CreatePermission(c *fiber.Ctx, tx *gorm.DB) (models.UserPermission, int, error) {

	var body models.UserPermission
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

	atdata := models.UserPermissionAt{
		RefId: body.UserId,
		UserPermissionContent: models.UserPermissionContent{
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

func UpdatePermission(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.UserPermission, int, error) {

	var body models.UserPermission

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := tx.Model(&models.UserPermission{}).
		Where("user_id = ?", body.UserId).
		Updates(map[string]interface{}{
			"can_create": body.CanCreate,
			"can_update": body.CanUpdate,
			"can_delete": body.CanDelete,
		}).Error; err != nil {

		return body, fiber.StatusInternalServerError, errors.New("failed updating user permission")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UserPermissionAt{
		RefId: body.UserId,
		UserPermissionContent: models.UserPermissionContent{
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

func DeletePermission(c *fiber.Ctx, tx *gorm.DB, id int) (int, error) {

	if err := services.DbDelete(tx, &models.UserPermissionAt{}, map[string]interface{}{"ref_id": id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user permission")
	}

	if err := services.DbDelete(tx, &models.UserPermission{}, map[string]interface{}{"id": id}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed deleting user permission")
	}

	return 0, nil
}
