package item_stock_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/utils"
)

type ItemStockHandler struct {
	Service *item_stock_services.ItemStockService
}

func NewItemStockHandler(service *item_stock_services.ItemStockService) *ItemStockHandler {
	return &ItemStockHandler{Service: service}
}

// GetItemStocksList backs the Inventory Item Stocks module's list screen - one row per
// item+warehouse+bin, with item/warehouse names already resolved.
func (h *ItemStockHandler) GetItemStocksList(c *fiber.Ctx) error {
	data, status, err := h.Service.GetItemStocksList()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetStockTransactions backs the stock ledger screen - one row per movement on
// tbl_inv_item_stocks (receiving, picking, requests, manual adds/adjustments, and any
// reversal of those), newest first. Pass ?item_id= to scope to one item.
func (h *ItemStockHandler) GetStockTransactions(c *fiber.Ctx) error {
	itemId, _ := strconv.Atoi(c.Query("item_id", "0"))

	data, status, err := h.Service.GetStockTransactions(uint(itemId))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetAvailableStock returns physical stock minus active quotation reservations, so a
// quotation screen can show what's actually free to promise rather than just what's
// physically in the warehouse. Pass ?item_id= to scope to one item.
func (h *ItemStockHandler) GetAvailableStock(c *fiber.Ctx) error {
	itemId, _ := strconv.Atoi(c.Query("item_id", "0"))

	data, status, err := h.Service.GetAvailableStock(uint(itemId))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// InsertItemStock is the "Add Stock" endpoint for the Inventory Item Stocks module - adds
// stock for an item+warehouse+bin combination. If a row for that exact combination already
// exists, the service upserts (adds the qty onto the existing row) instead of creating a
// duplicate - so this is safe to use whether or not the item already has stock there.
func (h *ItemStockHandler) InsertItemStock(c *fiber.Ctx) error {
	var body inventory_models.ItemStocks

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atBody := &inventory_models.ItemStocksAt{
		SourceType: "manual_add",
	}

	data, status, err := h.Service.InsertItemStock(&body, atBody, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// AdjustItemStock is the manual correction endpoint for the Inventory Item Stocks module -
// sets a bin's stock_qty directly to whatever was physically counted, audit-logged with
// the given Remarks.
func (h *ItemStockHandler) AdjustItemStock(c *fiber.Ctx) error {
	var body inventory_models.ItemStockAdjustmentBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.AdjustItemStock(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
