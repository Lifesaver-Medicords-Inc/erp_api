package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)


func GetUserPermissions(conditions map[string]interface{}) ([]adminmodels.UserPermission, int, error) {

	var positions []adminmodels.UserPermission

	if err := services.DbGet(&positions, conditions); err != nil {
		return positions, fiber.StatusInternalServerError, errors.New("failed getting positions")
	}

	return positions, 0, nil
}

func CreateUserPermission(c *fiber.Ctx, tx *gorm.DB) (adminmodels.UserPermission, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.UserPermission{}, adminmodels.UserPermissionAt{})

	var body adminmodels.UserPermission
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{RefId: body.ID, Code: body.UserId, At: at, UserPermission: adminmodels.UserPermission{
		UserId:    body.UserId,
		CanCreate: body.CanCreate,
		CanUpdate: body.CanUpdate,
		CanDelete: body.CanDelete,
	}}

	return service.Create(body, atdata)
}

func UpdateUserPermission(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.UserPermission, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.UserPermission{}, adminmodels.UserPermissionAt{})

	var body adminmodels.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{RefId: body.ID, Code: body.UserId, At: at, UserPermission: adminmodels.UserPermission{
		UserId:    body.UserId,
		CanCreate: body.CanCreate,
		CanUpdate: body.CanUpdate,
		CanDelete: body.CanDelete,
	}}

	return service.Update(body, atdata, conditions)
}

func DeleteUserPermission(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.UserPermission, int, error) {

	var body adminmodels.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.UserPermissionAt{RefId: body.ID, Code: body.UserId, At: at, UserPermission: adminmodels.UserPermission{
		UserId:    body.UserId,
		CanCreate: body.CanCreate,
		CanUpdate: body.CanUpdate,
		CanDelete: body.CanDelete,
	}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
