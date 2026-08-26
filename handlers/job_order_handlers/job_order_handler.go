package job_order_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/job_order_services"
	"github.com/pierceperado/smpc/utils"
)

type JobOrderHandler struct {
	Service *job_order_services.JobOrderService
}

func NewJobOrderHandler(service *job_order_services.JobOrderService) *JobOrderHandler {
	return &JobOrderHandler{Service: service}
}

func (h *JobOrderHandler) GetJobOrder(c *fiber.Ctx) error {
	idParam := c.Params("user_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"UserId": idNum,
	}

	data, status, err := h.Service.GetJobOrder(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetEngineerList(c *fiber.Ctx) error {
	data, status, err := h.Service.GetEngineerList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetSalesOrderViewEng(c *fiber.Ctx) error {
	data, status, err := h.Service.GetSalesOrderViewEng(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetComponents(c *fiber.Ctx) error {
	idParam := c.Params("bom_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"BomId": idNum,
	}
	data, status, err := h.Service.GetComponents(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) CreateJobOrder(c *fiber.Ctx) error {
	var body []models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateJobOrder(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) UpdateJobOrder(c *fiber.Ctx) error {
	var body []models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateJobOrder(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// AcceptJobOrder - §6.1 (D) "accept SO items for production", access-gated via
// JobOrderAcceptAccessCode.
func (h *JobOrderHandler) AcceptJobOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid job order id")
	}

	status, err := h.Service.AcceptJobOrder(uint(id), actingUserId(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// AcknowledgeJobOrder - §5.23's Warehouse Manager acknowledgement step, access-gated via
// JobOrderWhAckAccessCode. Body carries the destination the produced units go into
// stock at - see AcknowledgeJobOrder's own doc comment for why that's required rather
// than defaulted.
func (h *JobOrderHandler) AcknowledgeJobOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid job order id")
	}

	var body struct {
		WarehouseId uint   `json:"warehouse_id"`
		BinLocation string `json:"bin_location"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	status, err := h.Service.AcknowledgeJobOrder(uint(id), actingUserId(c), body.WarehouseId, body.BinLocation, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// GetPendingProductionReports - §5.23's Warehouse Manager acknowledgement queue,
// company-wide (Phase 2 item 2.4).
func (h *JobOrderHandler) GetPendingProductionReports(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPendingProductionReports()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// actingUserId pulls the numeric user id off the same "at" audit context every other
// write endpoint relies on (see utils/at_util.go and the identical helper in
// sales_return_handlers/item_stock_handlers) - there's no separate authentication/
// session concept in this API beyond that.
func actingUserId(c *fiber.Ctx) uint {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		return 0
	}

	id, err := strconv.Atoi(at.AtUserId)
	if err != nil || id < 0 {
		return 0
	}

	return uint(id)
}
