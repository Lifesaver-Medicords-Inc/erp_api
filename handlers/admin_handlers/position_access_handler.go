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

type PositionAccessHandler struct {
	PositionAccessService *adminservices.PositionAccessService
}

func NewPositionAccessHandler(service *adminservices.PositionAccessService) *PositionAccessHandler {
	return &PositionAccessHandler{
		PositionAccessService: service,
	}
}

func (p *PositionAccessHandler) GetAllPositionAccessHandler(c *fiber.Ctx) error {

	data, status, err := p.PositionAccessService.GetPositionAllAccessService(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionAccessHandler) GetPositionAccessHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}
	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := p.PositionAccessService.GetPositionAccessService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionAccessHandler) CreatePositionAccessHandler(c *fiber.Ctx) error {

	var body models.PositionAccessModel

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := p.PositionAccessService.CreatePositionAccessService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionAccessHandler) UpdatePositionAccessHandler(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.PositionAccessModel

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := p.PositionAccessService.UpdatePositionAccessService(&body, conditions, at)
	if err != nil {

		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (p *PositionAccessHandler) DeletePositionAccessHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := p.PositionAccessService.DeletePositionAccessService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionAccessHandler) UpdatePositionAllAccessHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	var accessList []models.PositionAccessModel

	if err := c.BodyParser(&accessList); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")

	}

	if len(accessList) == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Access list is empty")

	}

	tx := initializers.DB.Begin()

	// Delete PositionAccess records for the position
	if err := tx.Where("position_id = ?", idNum).Delete(&models.PositionAccessModel{}).Error; err != nil {
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
