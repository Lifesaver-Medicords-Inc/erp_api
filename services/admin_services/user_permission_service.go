package adminservices

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type PermissionService struct {
}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (p *PermissionService) GetPermissionService(conditions map[string]interface{}) (*models.UserPermissionModel, int, error) {

	var permissions = &models.UserPermissionModel{}

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, 404, errors.New("failed getting user permission")
	}

	return permissions, 200, nil
}

func (p *PermissionService) GetPermissionsService(conditions map[string]interface{}) (*[]models.UserPermissionModel, int, error) {

	var permissions = &[]models.UserPermissionModel{}

	if err := services.DbGetNoCache(&permissions, conditions); err != nil {
		return permissions, 404, errors.New("failed getting user permission")
	}

	return permissions, 200, nil
}

func (p *PermissionService) CreatePermissionService(permission *models.UserPermissionModel, at models.At) (*models.UserPermissionModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.UserPermissionModel{}, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &permission); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating user permission")
		}
		tx.Rollback()
		return permission, 500, err
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
		return permission, 500, errors.New("failed creating user permissionnat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, 500, errors.New("failed to commit transaction")
	}

	return permission, 200, nil
}

func (p *PermissionService) UpdatePermissionService(permission *models.UserPermissionModel, conditions map[string]interface{}, at models.At) (*models.UserPermissionModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.UserPermissionModel{}, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Model(&models.UserPermissionModel{}).
		Where("user_id = ?", permission.UserId).
		Updates(map[string]interface{}{
			"can_create": permission.CanCreate,
			"can_update": permission.CanUpdate,
			"can_delete": permission.CanDelete,
		}).Error; err != nil {

		return permission, 500, errors.New("failed updating user permission")
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
		return permission, 500, errors.New("failed creating user permissionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, 500, errors.New("failed to commit transaction")
	}

	return permission, 200, nil
}

func (p *PermissionService) DeletePermissionService(conditions map[string]interface{}, at models.At) (*models.UserPermissionModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.UserPermissionModel{}, 500, errors.New("failed to start DB transaction")
	}

	permission, status, err := p.GetPermissionService(conditions)

	if err != nil {
		return permission, status, errors.New("user not found")
	}

	if err := services.DbDelete(tx, &permission, conditions); err != nil {
		tx.Rollback()
		return permission, 500, errors.New("failed deleting permission")
	}
	atdata := models.UserPermissionAt{RefId: permission.ID, UserPermissionContent: permission.UserPermissionContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return permission, 500, errors.New("failed creating user_permisionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return permission, 500, errors.New("failed to commit transaction")
	}

	return permission, 200, nil
}
