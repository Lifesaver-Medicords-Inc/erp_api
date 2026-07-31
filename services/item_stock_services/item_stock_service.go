package item_stock_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type ItemStockService struct{}

func NewItemStockService() *ItemStockService {
	return &ItemStockService{}
}

func (s *ItemStockService) InsertItemStock(body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, at models.At) (*inventory_models.ItemStocks, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	result, err := s.UpsertStockWithTx(tx, body, atBody, at)
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateCaches()
	return result, fiber.StatusOK, nil
}

func (s *ItemStockService) UpdateItemStock(body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, conditions map[string]interface{}, at models.At) (*inventory_models.ItemStocks, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var existing inventory_models.ItemStocks
	err := tx.Where(
		"item_id = ? AND warehouse_id = ? AND bin_location = ?",
		body.ItemId, body.WarehouseId, body.BinLocation,
	).First(&existing).Error

	if err != nil {
		return body, fiber.StatusNotFound, errors.New("no stock record found for this item, warehouse, and bin combination")
	}

	if *body.StockQty > *existing.StockQty {
		return body, fiber.StatusUnprocessableEntity, fmt.Errorf(
			"insufficient stock: requested %d but only %d available in bin %s",
			body.StockQty, existing.StockQty, body.BinLocation,
		)
	}

	*existing.StockQty -= *body.StockQty
	s.SetActiveStatus(&existing)

	if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed updating item stocks")
	}

	atdata := inventory_models.ItemStocksAt{
		RefId:             existing.ID,
		SourceId:          atBody.SourceId,
		SourceType:        atBody.SourceType,
		ItemStocksContent: existing.ItemStocksContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item stocks at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateCaches()
	return &existing, fiber.StatusOK, nil
}

// UpsertStockWithTx is the shared core upsert logic that runs inside an existing transaction.
func (s *ItemStockService) UpsertStockWithTx(tx *gorm.DB, body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, at models.At) (*inventory_models.ItemStocks, error) {

	var existing inventory_models.ItemStocks

	err := tx.Where(
		"item_id = ? AND warehouse_id = ? AND bin_location = ?",
		body.ItemId, body.WarehouseId, body.BinLocation,
	).First(&existing).Error

	if err == nil {
		// Row exists — accumulate incoming qty into existing stock
		*existing.StockQty += *body.StockQty
		s.SetActiveStatus(&existing)

		if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
			return nil, errors.New("failed updating existing item stocks")
		}

		atdata := inventory_models.ItemStocksAt{
			RefId:             existing.ID,
			SourceId:          atBody.SourceId,
			SourceType:        atBody.SourceType,
			ItemStocksContent: existing.ItemStocksContent,
			At:                at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			return nil, errors.New("failed creating item stocks at")
		}

		return &existing, nil
	}

	// Row does not exist — fresh insert
	nextDocNo, err := utils.NextDocNo(tx, new(inventory_models.ReceivingReport), "doc_no")
	if err != nil {
		return nil, errors.New("failed getting next doc number")
	}

	body.DocNo = nextDocNo
	s.SetActiveStatus(body)

	if err := services.DbInsert(tx, body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("duplicate record error")
		}
		return nil, errors.New("failed creating item stocks")
	}

	atdata := inventory_models.ItemStocksAt{
		RefId:             body.ID,
		SourceId:          atBody.SourceId,
		SourceType:        atBody.SourceType,
		ItemStocksContent: body.ItemStocksContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, errors.New("failed creating item stocks at")
	}

	return body, nil
}

// DeductStockWithTx is the shared core upsert logic that runs inside an existing transaction.
func (s *ItemStockService) DeductStockWithTx(tx *gorm.DB, body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, at models.At) (*inventory_models.ItemStocks, error) {

	var existing inventory_models.ItemStocks

	err := tx.Where("id = ?", body.ID).First(&existing).Error

	//If no record exists → cannot deduct
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no stock record found for deduction")
		}
		return nil, errors.New("failed fetching item stock")
	}

	//Prevent nil pointer issues
	if existing.StockQty == nil || body.StockQty == nil {
		return nil, errors.New("invalid stock quantity")
	}

	//Check for insufficient stock
	if *existing.StockQty < *body.StockQty {
		return nil, fmt.Errorf(
			"insufficient stock: requested %d but only %d available in bin %s",
			*body.StockQty, *existing.StockQty, existing.BinLocation,
		)
	}

	//Deduct stock
	*existing.StockQty -= *body.StockQty
	s.SetActiveStatus(&existing)

	//Update DB
	if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
		return nil, errors.New("failed updating item stocks")
	}

	//Audit trail (same as your pattern)
	atdata := inventory_models.ItemStocksAt{
		RefId:             existing.ID,
		SourceId:          atBody.SourceId,
		SourceType:        atBody.SourceType,
		ItemStocksContent: existing.ItemStocksContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, errors.New("failed creating item stocks at")
	}

	return &existing, nil
}

// RestoreStockWithTx reverses a prior deduction — adds qty back to the bin
// identified by body.ID and writes an audit trail entry.
func (s *ItemStockService) RestoreStockWithTx(tx *gorm.DB, body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, at models.At) (*inventory_models.ItemStocks, error) {

	var existing inventory_models.ItemStocks

	if err := tx.Where("id = ?", body.ID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no stock record found for restoration")
		}
		return nil, errors.New("failed fetching item stock for restoration")
	}

	if existing.StockQty == nil || body.StockQty == nil {
		return nil, errors.New("invalid stock quantity for restoration")
	}

	*existing.StockQty += *body.StockQty
	s.SetActiveStatus(&existing)

	if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
		return nil, errors.New("failed restoring item stock")
	}

	atdata := inventory_models.ItemStocksAt{
		RefId:             existing.ID,
		SourceId:          atBody.SourceId,
		SourceType:        atBody.SourceType,
		ItemStocksContent: existing.ItemStocksContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, errors.New("failed creating item stocks at for restoration")
	}

	return &existing, nil
}

// GetItemStocksList returns every tbl_inv_item_stocks row (one per item+warehouse+bin),
// joined with item code/name/brand and warehouse name so the Inventory Item Stocks module
// (and any other caller, e.g. Sales Order's stock check) doesn't have to re-resolve IDs
// itself. No existing DB view already does this join against the live table - the
// pre-existing inventory views are built on the separate legacy tbl_inv_stocks_location
// table, which the current Receiving Report flow doesn't write to.
func (s *ItemStockService) GetItemStocksList() ([]inventory_models.ItemStockListView, int, error) {
	var response []inventory_models.ItemStockListView

	query := `
		SELECT its.id, its.item_id, b.item_code,
		       ISNULL(c.name, '') AS item_name,
		       ISNULL(d.name, '') AS brand,
		       its.warehouse_id, ISNULL(w.name, '') AS warehouse_name,
		       its.bin_location, its.stock_qty, its.stock_uom, its.is_active
		FROM tbl_inv_item_stocks its
		LEFT JOIN tbl_setup_item b ON its.item_id = b.id
		LEFT JOIN tbl_setup_item_name c ON b.item_name_id = c.id
		LEFT JOIN tbl_setup_item_brand d ON b.item_brand_id = d.id
		LEFT JOIN tbl_inv_warehouse_name w ON its.warehouse_id = w.id
		ORDER BY ISNULL(c.name, ''), its.bin_location
	`

	if err := initializers.DB.Raw(query).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting item stocks list")
	}

	return response, fiber.StatusOK, nil
}

// AdjustItemStock is a manual correction, distinct from UpsertStockWithTx/DeductStockWithTx
// (which add/subtract a delta as part of a receiving/issuing transaction) - this SETS
// stock_qty directly to whatever the user physically counted, and always writes an audit
// entry (with Remarks) so the correction is traceable later.
func (s *ItemStockService) AdjustItemStock(body *inventory_models.ItemStockAdjustmentBody, at models.At) (*inventory_models.ItemStocks, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var existing inventory_models.ItemStocks
	if err := tx.Where("id = ?", body.ID).First(&existing).Error; err != nil {
		return nil, fiber.StatusNotFound, errors.New("no stock record found for this bin")
	}

	newQty := body.NewQty
	existing.StockQty = &newQty
	s.SetActiveStatus(&existing)

	if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed updating item stock")
	}

	atdata := inventory_models.ItemStocksAt{
		RefId:             existing.ID,
		SourceType:        "manual_adjustment",
		Remarks:           body.Remarks,
		ItemStocksContent: existing.ItemStocksContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed creating item stock audit record")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateCaches()
	return &existing, fiber.StatusOK, nil
}

// setActiveStatus sets IsActive to true when StockQty > 0, false when zero.
// Pointer bool ensures GORM persists false without treating it as a zero-value omission.
func (s *ItemStockService) SetActiveStatus(stock *inventory_models.ItemStocks) {
	active := stock.StockQty != nil && *stock.StockQty > 0
	stock.IsActive = &active
}

func invalidateCaches() {
	setup_services.InvalidateItemCaches()
	if err := services.InvalidateCacheByModel(inventory_models.WarehouseReceivingAreaView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}
}
