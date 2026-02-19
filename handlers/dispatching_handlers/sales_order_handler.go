package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type SalesOrderHandler struct {
	Service *dispatching_services.SalesOrderService
}

func NewSalesOrderHandler(service *dispatching_services.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{Service: service}
}

// func (h *SalesOrderHandler) GetSalesOrdersHandler(c *fiber.Ctx) error {
// 	id := c.Query("id")
// 	customer := c.Query("customerName")
// 	status := c.Query("status")

// 	conditions := make(map[string]interface{})

// 	if idNum, _ := strconv.Atoi(id); idNum != 0 {
// 		conditions["order_id"] = idNum
// 	}
// 	if customer != "" {
// 		conditions["customer_name"] = customer
// 	}
// 	if status != "" {
// 		conditions["status"] = status
// 	}

// 	data, code, err := h.Service.GetSalesOrdersService(conditions)
// 	if err != nil {
// 		return utils.RespondError(c, code, err.Error())
// 	}
// 	return utils.RespondSuccess(c, data)
// }

// func (h *SalesOrderHandler) GetSalesOrderHandler(c *fiber.Ctx) error {
// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
// 	}

// 	conditions := map[string]interface{}{"order_id": idNum}

// 	order, code, err := h.Service.GetSalesOrderService(conditions)
// 	if err != nil {
// 		return utils.RespondError(c, code, err.Error())
// 	}
// 	return utils.RespondSuccess(c, order)
// }

func (h *SalesOrderHandler) CreateSalesOrderHandler(c *fiber.Ctx) error {
	var body models.Order
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateSalesOrderService(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesOrderHandler) UpdateSalesOrderHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	var body models.Order
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	conditions := map[string]interface{}{"id": idNum}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateSalesOrderService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// func (h *SalesOrderHandler) DeleteSalesOrderHandler(c *fiber.Ctx) error {
// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
// 	}

// 	conditions := map[string]interface{}{"id": idNum}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	data, code, err := h.Service.DeleteSalesOrderService(conditions, at)
// 	if err != nil {
// 		return utils.RespondError(c, code, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }
