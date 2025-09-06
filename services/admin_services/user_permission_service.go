package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetPermission(conditions map[string]interface{}) (models.UserPermission, int, error) {

	var permissions models.UserPermission

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, fiber.StatusNotFound, errors.New("failed getting user permission")
	}

	return permissions, 0, nil
}

func GetPermissions(conditions map[string]interface{}) ([]models.UserPermission, int, error) {

	var permissions []models.UserPermission

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, fiber.StatusNotFound, errors.New("failed getting user permission")
	}

	return permissions, 0, nil
}

func CreatePermission(permission models.UserPermission, at models.At) (models.UserPermission, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.UserPermission{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &permission); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating user permission")
		}
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, err
	}

	atdata := models.UserPermissionAt{
		RefId: permission.UserId,
		UserPermissionContent: models.UserPermissionContent{
			UserId:    permission.UserId,
			CanCreate: permission.CanCreate,
			CanUpdate: permission.CanUpdate,
			CanDelete: permission.CanDelete,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed creating user permissionnat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return permission, 0, nil
}

func UpdatePermission(permission models.UserPermission, conditions map[string]interface{}, at models.At) (models.UserPermission, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.UserPermission{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Model(&models.UserPermission{}).
		Where("user_id = ?", permission.UserId).
		Updates(map[string]interface{}{
			"can_create": permission.CanCreate,
			"can_update": permission.CanUpdate,
			"can_delete": permission.CanDelete,
		}).Error; err != nil {

		return permission, fiber.StatusInternalServerError, errors.New("failed updating user permission")
	}

	atdata := models.UserPermissionAt{
		RefId: permission.UserId,
		UserPermissionContent: models.UserPermissionContent{
			UserId:    permission.UserId,
			CanCreate: permission.CanCreate,
			CanUpdate: permission.CanUpdate,
			CanDelete: permission.CanDelete,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed creating user permissionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return permission, 0, nil
}

func DeletePermission(conditions map[string]interface{}, at models.At) (models.UserPermission, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.UserPermission{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	permission, status, err := GetPermission(conditions)

	if err != nil {
		return permission, status, errors.New("user not found")
	}

	if err := services.DbDelete(tx, &permission, conditions); err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed deleting permission")
	}
	atdata := models.UserPermissionAt{RefId: permission.ID, UserPermissionContent: permission.UserPermissionContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed creating user_permisionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return permission, 0, nil
}
