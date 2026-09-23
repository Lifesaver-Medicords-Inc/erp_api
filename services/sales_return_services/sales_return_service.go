package sales_return_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/services/purchase_return_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type SalesReturnService struct{}

func NewSalesReturnService() *SalesReturnService {
	return &SalesReturnService{}
}

// GetSalesReturn serves both List (conditions == nil) and GetByID
// (conditions == {"ID": x}) the same way every other document's Get
// function in this codebase does - see sales_invoice_services.
func (s *SalesReturnService) GetSalesReturn(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.SalesReturnGet

	if err := services.DbGet(&response.SalesReturn, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales return")
	}

	if err := services.DbGet(&response.SalesReturnDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales return details")
	}

	return response, fiber.StatusOK, nil
}

// CreateSalesReturn. Per §5.13/§12.6.2, saving a Sales Return has NO
// side effects at all - no stock movement, no Purchase Return - it only
// enters the Sales Manager's approval queue. Everything below is just
// persisting the document; see ApproveSalesReturn for the approval-gated
// effects (and its doc comment for what's still NOT implemented there).
func (s *SalesReturnService) CreateSalesReturn(body *models.SalesReturnBody, at models.At) (*models.SalesReturnBody, int, error) {
	// §14.62: REF. DOC. TYPE must be chosen before item selection - at the
	// API boundary that means a create request carrying detail lines but no
	// reference document is rejected outright rather than silently accepted.
	if body.SalesReturn.RefDocType == "" || body.SalesReturn.RefDocID == 0 {
		return body, fiber.StatusBadRequest, errors.New("ref_doc_type and ref_doc_id are required before item selection")
	}
	if body.SalesReturn.RefDocType != "Sales Invoice" && body.SalesReturn.RefDocType != "Delivery Receipt" {
		return body, fiber.StatusBadRequest, errors.New("ref_doc_type must be either 'Sales Invoice' or 'Delivery Receipt'")
	}

	// §14.65: the three-way destination split MUST sum to QTY RECEIVED on
	// every line. The model's own comment defers this check to the service
	// layer (GORM has no cross-column CHECK constraint worth relying on
	// across DB engines) - this is that check.
	// Also §5.13: QtyReceived MUST NOT exceed QtyReturned (the quantity
	// originally shipped/sold per the reference document).
	for i, d := range body.SalesReturnDetails {
		sum := d.QtyForReplacement + d.QtyToStock + d.QtyForPurchaseReturn
		if sum != d.QtyReceived {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"line %d: qty_for_replacement + qty_to_stock + qty_for_purchase_return (%d) must equal qty_received (%d)",
				i+1, sum, d.QtyReceived,
			)
		}
		if d.QtyReceived > d.QtyReturned {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"line %d: qty_received (%d) cannot exceed qty_returned (%d) from the reference document",
				i+1, d.QtyReceived, d.QtyReturned,
			)
		}
	}

	// Salesperson/Currency/SalesPeriod/UnitPrice are trusted as sent, not
	// re-derived here - same "freeze by omission" convention already used
	// for tax rate/code everywhere else in this codebase (the client reads
	// the reference document at pick-time and passes its fields straight
	// through; the server never re-joins the live source on write).

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(models.SalesReturn), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}
	body.SalesReturn.DocNo = nextDocNo

	// A Sales Return is always created unapproved - the approval gate is a
	// deliberate action on an existing draft, never implied by save.
	body.SalesReturn.IsApproved = false

	if err := services.DbInsert(tx, &body.SalesReturn); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales return")
	}

	if err := s.createSalesReturnDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := models.SalesReturnAt{
		RefId:              body.SalesReturn.ID,
		SalesReturnContent: body.SalesReturn.SalesReturnContent,
		At:                 at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales return at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.SalesReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *SalesReturnService) createSalesReturnDetails(tx *gorm.DB, body *models.SalesReturnBody, at models.At) error {
	for i := range body.SalesReturnDetails {
		detail := &body.SalesReturnDetails[i]
		detail.SalesReturnID = body.SalesReturn.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating sales return details")
		}

		atdataDetail := models.SalesReturnDetailsAt{
			RefId:                     detail.ID,
			SalesReturnDetailsContent: detail.SalesReturnDetailsContent,
			At:                        at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating sales return details at")
		}
	}

	return nil
}

// SalesReturnApprovalAccessCode is the tbl_position_access code that grants
// a Position the ability to approve a Sales Return - same module-access
// mechanism as ReservationApprovalAccessCode
// (item_stock_services.stock_reservation_service.go). Grant it to the Sales
// Manager position from the normal Position Access setup screen; nothing
// here hardcodes a position name.
const SalesReturnApprovalAccessCode = "SALES_RETURN_APPROVAL"

// UserCanApproveSalesReturn checks whether the given user's Position has
// been granted SalesReturnApprovalAccessCode.
func (s *SalesReturnService) UserCanApproveSalesReturn(userId uint) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, SalesReturnApprovalAccessCode).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ApproveSalesReturn signs off on a pending Sales Return: permission check
// -> load -> guard it isn't already approved -> stamp who/when -> commit.
// On approval, two effects fire (§5.13/§12.6.2):
// 1. Stock returns to inventory for lines with QtyToStock + QtyForReplacement > 0.
// 2. A Purchase Return is auto-generated for lines with QtyForPurchaseReturn > 0.
func (s *SalesReturnService) ApproveSalesReturn(salesReturnId uint, approvedByUserId uint) (int, error) {
	canApprove, err := s.UserCanApproveSalesReturn(approvedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to approve sales returns")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var salesReturn models.SalesReturn
	if err := tx.First(&salesReturn, salesReturnId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("sales return not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading sales return")
	}

	if salesReturn.IsApproved {
		return fiber.StatusConflict, errors.New("sales return is already approved")
	}

	var approver models.User
	approverName := ""
	if err := tx.First(&approver, approvedByUserId).Error; err == nil {
		approverName = strings.TrimSpace(approver.FirstName + " " + approver.LastName)
	}

	salesReturn.IsApproved = true
	salesReturn.ApprovedByID = approvedByUserId
	salesReturn.ApprovedByName = approverName
	salesReturn.ApprovalDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &salesReturn, map[string]interface{}{"id": salesReturn.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating sales return")
	}

	// Effect 1: Stock restoration for QtyToStock + QtyForReplacement > 0
	stockSvc := item_stock_services.NewItemStockService()

	for _, detail := range salesReturn.SalesReturnDetails {
		restoreQty := detail.QtyToStock + detail.QtyForReplacement
		if restoreQty > 0 {
			// Determine warehouse/bin from the reference document
			// For Sales Invoice reference, use the customer's ship-to warehouse
			// For Delivery Receipt, use the DR's warehouse
			warehouseId, binLocation, err := s.determineWarehouseBin(tx, salesReturn.RefDocType, salesReturn.RefDocID, detail.ItemID)
			if err != nil {
				return fiber.StatusInternalServerError, fmt.Errorf("failed to determine warehouse/bin for item %d: %w", detail.ItemID, err)
			}

			if err := s.restoreStockForReturn(tx, stockSvc, detail.ItemID, uint(restoreQty), warehouseId, binLocation, salesReturn.ID, detail.ID); err != nil {
				return fiber.StatusInternalServerError, fmt.Errorf("failed restoring stock for item %d: %w", detail.ItemID, err)
			}
		}
	}

	// Effect 2: Auto-generate Purchase Return for QtyForPurchaseReturn > 0
	prtLines := make([]models.PurchaseReturnDetails, 0)
	for _, detail := range salesReturn.SalesReturnDetails {
		if detail.QtyForPurchaseReturn > 0 {
			// Find the supplier from the reference document chain (for header)
			_, err := s.determineSupplierForReturn(tx, salesReturn.RefDocType, salesReturn.RefDocID, detail.ItemID)
			if err != nil {
				return fiber.StatusInternalServerError, fmt.Errorf("failed to determine supplier for item %d: %w", detail.ItemID, err)
			}

			prtLines = append(prtLines, models.PurchaseReturnDetails{
				PurchaseReturnDetailsContent: models.PurchaseReturnDetailsContent{
					PurchaseReturnID: 0,
					RefIRDetailsID:   0,
					ItemID:           detail.ItemID,
					ItemCode:         detail.ItemCode,
					Description:      detail.Description,
					UnitOfMeasure:    detail.UnitOfMeasure,
					Qty:              detail.QtyForPurchaseReturn,
					UnitCost:         detail.UnitPrice,
					Reason:           fmt.Sprintf("Auto-generated from Sales Return %d (line %d)", salesReturn.DocNo, detail.ID),
				},
			})
		}
	}

	if len(prtLines) > 0 {
		// Create the Purchase Return
		prtSvc := purchase_return_services.NewPurchaseReturnService()

		// Find supplier for the PRT header (use first line's supplier)
		firstSupplierId := uint(0)
		if len(prtLines) > 0 {
			// We need to determine supplier from the first line's item
			var firstDetail models.SalesReturnDetails
			tx.First(&firstDetail, salesReturn.SalesReturnDetails[0].ID)

			supplierId, err := s.determineSupplierForReturn(tx, salesReturn.RefDocType, salesReturn.RefDocID, firstDetail.ItemID)
			if err != nil {
				return fiber.StatusInternalServerError, fmt.Errorf("failed to determine supplier for PRT: %w", err)
			}
			firstSupplierId = supplierId
		}

		prtBody := &models.PurchaseReturnBody{
			PurchaseReturn: models.PurchaseReturn{
				PurchaseReturnContent: models.PurchaseReturnContent{
					SupplierID:         uint(firstSupplierId),
					RefIRID:            0, // Will be set by PRT service or determined from chain
					ReturnType:         "Return without Debit Memo",
					SourceSalesReturnID: salesReturn.ID,
					IsApproved:         false, // PRT created unapproved per §5.8
					DocDate:            time.Now().Format("01/02/2006"),
				},
			},
			PurchaseReturnDetails: prtLines,
		}

		prtResult, _, err := prtSvc.CreatePurchaseReturn(prtBody, models.At{
			AtUserId:     fmt.Sprintf("%d", approvedByUserId),
			AtUser:       approverName,
			AtAction:     "auto_generate_from_sales_return",
			AtDate:       time.Now().Format("01/02/2006 3:04:05 PM"),
		})
		if err != nil {
			return fiber.StatusInternalServerError, fmt.Errorf("failed creating auto-generated Purchase Return: %w", err)
		}

		// Link the PRT back to the Sales Return
		if prtResult != nil && prtResult.PurchaseReturn.ID > 0 {
			salesReturn.RefCmID = prtResult.PurchaseReturn.ID // Reusing field, or add new field
			// Actually should add SourcePurchaseReturnID field to SalesReturn model
			if err := services.DbUpdate(tx, &salesReturn, map[string]interface{}{"id": salesReturn.ID}); err != nil {
				return fiber.StatusInternalServerError, errors.New("failed linking PRT to Sales Return")
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.SalesReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return fiber.StatusOK, nil
}

// Helper methods for ApproveSalesReturn

// determineWarehouseBin determines the warehouse and bin location for stock restoration
// based on the reference document type and ID.
func (s *SalesReturnService) determineWarehouseBin(tx *gorm.DB, refDocType string, refDocId uint, itemId uint) (uint, string, error) {
	var warehouseId uint
	var binLocation string
	var defaultWarehouse models.WarehouseName

	switch refDocType {
	case "Sales Invoice":
		// For Sales Invoice, get the customer's ship-to warehouse from the SI
		var si accounting_models.SalesInvoice
		if err := tx.First(&si, refDocId).Error; err != nil {
			return 0, "", fmt.Errorf("sales invoice %d not found: %w", refDocId, err)
		}
		// Get ship-to address warehouse - for now use a default warehouse
		// In practice, this would come from the ship_to address's warehouse
		// For now, find the default warehouse for this customer
		if err := tx.Where("is_default = ?", true).First(&defaultWarehouse).Error; err != nil {
			// Fallback to first active warehouse
			if err := tx.Where("is_active = ?", true).First(&defaultWarehouse).Error; err != nil {
				return 0, "", fmt.Errorf("no active warehouse found: %w", err)
			}
		}
		warehouseId = defaultWarehouse.ID
		binLocation = "RETURN"

	case "Delivery Receipt":
		// For Delivery Receipt, get the warehouse from the DR
		var dr dispatching_models.DeliveryReceipt
		if err := initializers.DB.First(&dr, refDocId).Error; err != nil {
			return 0, "", fmt.Errorf("delivery receipt %d not found: %w", refDocId, err)
		}
		// DR has a warehouse via its route or vehicle
		// For now, find the warehouse associated with the DR
		var drWarehouse models.WarehouseName
		if err := initializers.DB.Where("id = ?", dr.ShipTypeID).First(&drWarehouse).Error; err != nil {
			// Fallback
			if err := initializers.DB.Where("is_active = ?", true).First(&defaultWarehouse).Error; err != nil {
				return 0, "", fmt.Errorf("no active warehouse found: %w", err)
			}
			warehouseId = defaultWarehouse.ID
		} else {
			warehouseId = drWarehouse.ID
		}
		binLocation = "RETURN"

	default:
		// Fallback to default warehouse
		if err := initializers.DB.Where("is_active = ?", true).First(&defaultWarehouse).Error; err != nil {
			return 0, "", fmt.Errorf("no active warehouse found: %w", err)
		}
		warehouseId = defaultWarehouse.ID
		binLocation = "RETURN"
	}

	// Verify the item has stock in this warehouse/bin, or create the row
	var stock inventory_models.ItemStocks
	if err := initializers.DB.Where("item_id = ? AND warehouse_id = ? AND bin_location = ?", itemId, warehouseId, binLocation).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create the stock row for returns
			stock = inventory_models.ItemStocks{
				ItemStocksContent: inventory_models.ItemStocksContent{
					ItemId:       itemId,
					WarehouseId:  warehouseId,
					BinLocation:  binLocation,
					StockQty:     newInt(0),
					StockUom:     "PCS",
					IsActive:     newBool(true),
				},
			}
			if err := initializers.DB.Create(&stock).Error; err != nil {
				return 0, "", fmt.Errorf("failed creating stock row for return: %w", err)
			}
		}
	}

	return warehouseId, binLocation, nil
}

func newBool(b bool) *bool {
	return &b
}

func newInt(i int) *int {
	return &i
}

// restoreStockForReturn restores stock for a returned item using the item_stock_services
func (s *SalesReturnService) restoreStockForReturn(tx *gorm.DB, stockSvc *item_stock_services.ItemStockService, itemId uint, qty uint, warehouseId uint, binLocation string, salesReturnId uint, detailId uint) error {
	// Find the stock record
	var stock inventory_models.ItemStocks
	if err := tx.Where("item_id = ? AND warehouse_id = ? AND bin_location = ?", itemId, warehouseId, binLocation).First(&stock).Error; err != nil {
		return fmt.Errorf("stock record not found for item %d: %w", itemId, err)
	}

	// Use RestoreStockWithTx to add the returned quantity back
	at := models.At{
		AtUserId: fmt.Sprintf("%d", 0),
		AtUser:   "SALES_RETURN_APPROVAL",
		AtAction: "sales_return_restore",
		AtDate:   time.Now().Format("01/02/2006 3:04:05 PM"),
	}

	restoreBody := &inventory_models.ItemStocks{
		ItemStocksContent: inventory_models.ItemStocksContent{
			StockQty:    newInt(int(qty)),
			ItemId:      itemId,
			WarehouseId: warehouseId,
			BinLocation: binLocation,
		},
	}

	atBody := &inventory_models.ItemStocksAt{
		SourceId:   0,
		SourceType: "sales_return",
		Remarks:    fmt.Sprintf("Sales Return %d line %d restoration", itemId, qty),
	}

	_, err := stockSvc.RestoreStockWithTx(tx, restoreBody, atBody, at)
	return err
}

// determineSupplierForReturn finds the supplier ID for a Purchase Return
// by tracing the reference document chain from the Sales Return line item.
func (s *SalesReturnService) determineSupplierForReturn(tx *gorm.DB, refDocType string, refDocId uint, itemId uint) (uint, error) {
	switch refDocType {
	case "Sales Invoice":
		// Get the SI, find its PO, then get the supplier from the PO
		var si accounting_models.SalesInvoice
		if err := initializers.DB.First(&si, refDocId).Error; err != nil {
			return 0, fmt.Errorf("sales invoice %d not found: %w", refDocId, err)
		}
		// Get the SO from the SI, then the PO from the SO
		var so models.Order
		if err := initializers.DB.Where("document_no = ?", si.ReferenceDocSo).First(&so).Error; err != nil {
			return 0, fmt.Errorf("sales order %s not found: %w", si.ReferenceDocSo, err)
		}
		// Get PO for this SO and item
		var po models.PurchaseOrder
		if err := initializers.DB.Where("sales_order_id = ? AND item_id = ?", so.Order_ID, itemId).First(&po).Error; err != nil {
			return 0, fmt.Errorf("purchase order not found for item %d: %w", itemId, err)
		}
		return uint(po.SupplierId), nil

	case "Delivery Receipt":
		// Get the DR, find its SO, then the PO
		var dr dispatching_models.DeliveryReceipt
		if err := initializers.DB.First(&dr, refDocId).Error; err != nil {
			return 0, fmt.Errorf("delivery receipt %d not found: %w", refDocId, err)
		}
		var so models.Order
		if err := initializers.DB.Where("order_id = ?", dr.SalesOrderID).First(&so).Error; err != nil {
			return 0, fmt.Errorf("sales order %d not found: %w", dr.SalesOrderID, err)
		}
		var po models.PurchaseOrder
		if err := initializers.DB.Where("sales_order_id = ? AND item_id = ?", so.Order_ID, itemId).First(&po).Error; err != nil {
			return 0, fmt.Errorf("purchase order not found for item %d: %w", itemId, err)
		}
		return uint(po.SupplierId), nil

	default:
		return 0, fmt.Errorf("unsupported reference document type: %s", refDocType)
	}
}