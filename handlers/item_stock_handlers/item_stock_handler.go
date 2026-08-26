package item_stock_handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/utils"
)

// CreateStockReservationBody is what the stock-check modal's RESERVE checkbox posts.
// ExpiresAt is optional - the WinForms client sends the quotation's own ValidUntil
// (already known client-side) formatted as "2006-01-02" or RFC3339; left blank, the
// reservation never gets picked up by the periodic expiry sweep and needs manual
// cleanup (same limitation the old auto-reserve path had).
type CreateStockReservationBody struct {
	ItemId      uint   `json:"item_id"`
	Qty         uint   `json:"qty"`
	SourceType  string `json:"source_type"`
	SourceId    uint   `json:"source_id"`
	QuotationId uint   `json:"quotation_id"`
	ExpiresAt   string `json:"expires_at"`
}

func parseReservationExpiry(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	for _, layout := range []string{"2006-01-02T15:04:05Z07:00", "2006-01-02"} {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return &t
		}
	}

	return nil
}

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

// GetStockReservation reports whether one quotation line currently has a manual
// reservation (see CreateStockReservation/ReleaseStockReservation below) - the
// stock-check modal calls this per line when it opens, to draw RESERVE checked or not.
// Pass ?source_type= (defaults to "sales_quotation") and ?source_id=. Responds with
// null data if nothing's reserved for that line, not an error.
func (h *ItemStockHandler) GetStockReservation(c *fiber.Ctx) error {
	sourceType := c.Query("source_type", "sales_quotation")
	sourceId, _ := strconv.Atoi(c.Query("source_id", "0"))

	if sourceId == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "source_id is required")
	}

	reservation, err := h.Service.GetReservation(sourceType, uint(sourceId))
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed getting stock reservation")
	}

	return utils.RespondSuccess(c, reservation)
}

// CreateStockReservation lets a sales rep/manager manually check "RESERVE" on the
// stock-check modal in Quick Quote / Project Quotation - the only place a reservation
// gets placed from the UI now (creating/editing a quotation line no longer reserves
// stock on its own; see quick_quotation_service.go).
func (h *ItemStockHandler) CreateStockReservation(c *fiber.Ctx) error {
	var body CreateStockReservationBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if body.ItemId == 0 || body.SourceId == 0 || body.Qty == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "item_id, source_id and qty are required")
	}

	sourceType := body.SourceType
	if sourceType == "" {
		sourceType = "sales_quotation"
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	status, err := h.Service.CreateStockReservationByRef(
		body.ItemId, body.Qty, sourceType, body.SourceId, body.QuotationId, parseReservationExpiry(body.ExpiresAt), at.AtUser,
	)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// ReleaseStockReservation lets a sales manager manually release a quotation line's soft
// stock hold early (the "RESERVE" checkbox on the stock-check modal in Quick Quote /
// Project Quotation) - freeing it back into the shared pool for other quotations.
// Pass ?source_type= (defaults to "sales_quotation") and ?source_id= (the
// SalesQuotationQuick line's id).
func (h *ItemStockHandler) ReleaseStockReservation(c *fiber.Ctx) error {
	sourceType := c.Query("source_type", "sales_quotation")
	sourceId, _ := strconv.Atoi(c.Query("source_id", "0"))

	if sourceId == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "source_id is required")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	status, err := h.Service.ReleaseStockReservationByRef(sourceType, uint(sourceId), at.AtUser)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// GetPendingReservations backs the dispatcher/inventory manager's approval queue -
// every stock reservation still awaiting a decision, oldest first.
func (h *ItemStockHandler) GetPendingReservations(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPendingReservations()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// actingUserId pulls the numeric user id off the same "at" audit context every other
// write endpoint already relies on (see utils/at_util.go) - there's no separate
// authentication/session concept in this API beyond that, so it's what identifies who's
// clicking Approve/Reject for the position-access check in the service layer.
func actingUserId(c *fiber.Ctx) uint {
	// TEMP DEBUG - remove once RESERVATION_APPROVAL 403s are confirmed fixed.
	atRaw := c.Locals("at")
	at, ok := atRaw.(models.At)
	if !ok {
		fmt.Printf("[RESV-DEBUG] c.Locals(\"at\") missing or wrong type - RequireAuth did not run on this request. raw=%#v\n", atRaw)
		return 0
	}
	fmt.Printf("[RESV-DEBUG] c.Locals(\"at\") = %#v\n", at)

	id, err := strconv.Atoi(at.AtUserId)
	if err != nil || id < 0 {
		fmt.Printf("[RESV-DEBUG] AtUserId %q did not parse to a valid uint: %v\n", at.AtUserId, err)
		return 0
	}

	fmt.Printf("[RESV-DEBUG] actingUserId resolved to %d\n", id)
	return uint(id)
}

// ApproveReservation signs off on a pending reservation. Only a user whose Position has
// been granted the RESERVATION_APPROVAL access code (see
// item_stock_services.ReservationApprovalAccessCode) can do this - anyone else gets a
// 403, checked server-side rather than trusted to the client hiding the button.
func (h *ItemStockHandler) ApproveReservation(c *fiber.Ctx) error {
	reservationId, err := strconv.Atoi(c.Params("id"))
	if err != nil || reservationId <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid reservation id is required")
	}

	status, err := h.Service.ApproveReservation(uint(reservationId), actingUserId(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// RejectReservation declines a pending reservation and frees its held stock back up.
// Same RESERVATION_APPROVAL gate as ApproveReservation.
func (h *ItemStockHandler) RejectReservation(c *fiber.Ctx) error {
	reservationId, err := strconv.Atoi(c.Params("id"))
	if err != nil || reservationId <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid reservation id is required")
	}

	status, err := h.Service.RejectReservation(uint(reservationId), actingUserId(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
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

	data, status, err := h.Service.AdjustItemStock(&body, actingUserId(c), at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// TransferStock is §10.6's "Transfer" function - move stock between bins/warehouses
// with no reference document, gated to Admin and the Warehouse Manager (§14.87) via the
// same StockTransferAccessCode as AdjustItemStock.
func (h *ItemStockHandler) TransferStock(c *fiber.Ctx) error {
	var body inventory_models.StockTransferBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	status, err := h.Service.TransferStock(&body, actingUserId(c), at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}
