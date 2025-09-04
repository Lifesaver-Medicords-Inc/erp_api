package adminhandlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

func GetAllPositionAccess(c *fiber.Ctx) error {

	tx := initializers.DB.Begin()
	data, status, err := adminservices.GetPositionAccess(nil, tx)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetPositionAccess(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}
	conditions := map[string]interface{}{
		"id": idNum,
	}
	tx := initializers.DB.Begin()
	data, status, err := adminservices.GetPositionAccess(conditions, tx)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data[0])
}

func CreatePositionAccess(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.CreatePositionAccess(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdatePositionAccess(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.UpdatePositionAccess(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func DeletePositionAccess(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.DeletePositionAccess(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdatePositionAllAccess(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	var accessList []models.PositionAccess

	if err := c.BodyParser(&accessList); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")

	}

	if len(accessList) == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Access list is empty")

	}

	tx := initializers.DB.Begin()

	var accessIDs []int
	if err := tx.Model(&models.PositionAccess{}).
		Where("position_id = ?", idNum).
		Pluck("id", &accessIDs).Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to fetch existing access IDs")
	}

	// Delete all PositionAccessAt where ref_id in accessIDs
	if len(accessIDs) > 0 {
		if err := tx.Where("ref_id IN ?", accessIDs).Delete(&models.PositionAccessAt{}).Error; err != nil {
			tx.Rollback()
			return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to delete existing PositionAccessAt")
		}
	}

	// Delete PositionAccess records for the position
	if err := tx.Where("position_id = ?", idNum).Delete(&models.PositionAccess{}).Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to delete existing PositionAccess")
	}

	for _, a := range accessList {

		if err := services.DbInsert(tx, &a); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				tx.Rollback()
				return errors.New("duplicate record error")
			} else {
				tx.Rollback()
				return errors.New("failed creating position access")
			}
		}

		at, ok := c.Locals("at").(models.At)
		if !ok {
			at = models.At{}
		}

		atdata := models.PositionAccessAt{
			RefId: a.ID,
			Code:  a.Code,
			PositionAccessContent: models.PositionAccessContent{
				PositionId: a.PositionId,
				Code:       a.Code,
			},
			At: at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return utils.RespondSuccess(c, accessList)
}
