package adminhandlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type PositionWithAccess struct {
	models.PositionModel
	Access []*models.PositionAccessModel `json:"access"`
}

type PositionHandler struct {
	PositionService *adminservices.PositionService
}

func NewPositionHandler(service *adminservices.PositionService) *PositionHandler {
	return &PositionHandler{
		PositionService: service,
	}
}

func (p *PositionHandler) GetPositions(c *fiber.Ctx) error {

	id := c.Query("id")
	name := c.Query("name")
	code := c.Query("code")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)

	if idNum != 0 {
		conditions["id"] = id
	}

	if name != "" {
		conditions["name"] = name
	}

	if code != "" {
		conditions["code"] = code
	}

	positions, status, err := p.PositionService.GetPositions(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, positions)
}

func (p *PositionHandler) GetPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	position, status, err := p.PositionService.GetPosition(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, position)
}

func (p *PositionHandler) CreatePosition(c *fiber.Ctx) error {

	var body models.PositionModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := p.PositionService.CreatePosition(body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionHandler) UpdatePosition(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.PositionModel
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

	data, status, err := p.PositionService.UpdatePosition(body, conditions, at)
	if err != nil {

		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (p *PositionHandler) DeletePosition(c *fiber.Ctx) error {

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

	data, status, err := p.PositionService.DeletePosition(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
