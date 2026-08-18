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

	result, err := s.UpsertStockWithTx(tx, body, atBody, at, nil)
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

	if err := services.SetStockAuditContext(tx, atBody.SourceType, atBody.SourceId, atBody.Remarks, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed setting stock audit context")
	}

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

// UpsertStockWithTx is the shared core upsert logic that runs inside an existing
// transaction. lot is optional (pass nil for callers that aren't a real purchase, e.g.
// manual stock add) - when non-nil, a new StockLot row is created for the incoming qty
// so a later sale can draw from it FIFO (see ConsumeLotsFIFO), and the same cost info
// is attached to this IN movement's ledger row via SetStockAuditContext.
func (s *ItemStockService) UpsertStockWithTx(tx *gorm.DB, body *inventory_models.ItemStocks, atBody *inventory_models.ItemStocksAt, at models.At, lot *inventory_models.LotInfo) (*inventory_models.ItemStocks, error) {

	var existing inventory_models.ItemStocks

	err := tx.Where(
		"item_id = ? AND warehouse_id = ? AND bin_location = ?",
		body.ItemId, body.WarehouseId, body.BinLocation,
	).First(&existing).Error

	if err == nil {
		// Row exists — accumulate incoming qty into existing stock
		*existing.StockQty += *body.StockQty
		s.SetActiveStatus(&existing)

		if err := services.SetStockAuditContext(tx, atBody.SourceType, atBody.SourceId, atBody.Remarks, lot); err != nil {
			return nil, errors.New("failed setting stock audit context")
		}

		if lot != nil {
			if err := s.CreateStockLot(tx, existing.ItemId, existing.WarehouseId, existing.BinLocation, *body.StockQty, lot); err != nil {
				return nil, fmt.Errorf("failed creating stock lot: %w", err)
			}
		}

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

	if err := services.SetStockAuditContext(tx, atBody.SourceType, atBody.SourceId, atBody.Remarks, lot); err != nil {
		return nil, errors.New("failed setting stock audit context")
	}

	if err := s.insertItemStockNoOutput(tx, body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("duplicate record error")
		}
		return nil, errors.New("failed creating item stocks")
	}

	if lot != nil {
		if err := s.CreateStockLot(tx, body.ItemId, body.WarehouseId, body.BinLocation, *body.StockQty, lot); err != nil {
			return nil, fmt.Errorf("failed creating stock lot: %w", err)
		}
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

// insertItemStockNoOutput inserts a brand-new tbl_inv_item_stocks row, bypassing
// tx.Create(). GORM's sqlserver driver unconditionally appends "OUTPUT INSERTED.<col>"
// to every INSERT it builds for a model with an auto-increment primary key (see
// go-gorm/sqlserver's outputInserted() in create.go) - there's no per-model or per-call
// way to opt out of that. SQL Server itself rejects an OUTPUT clause without an INTO
// clause on any table that has an enabled trigger, and tr_inv_item_stocks_ledger.sql
// (sql/triggers/tr_inv_item_stocks_ledger.sql) is exactly such a trigger on this table,
// so every tx.Create() against tbl_inv_item_stocks fails with:
//
//	mssql: The target table 'tbl_inv_item_stocks' of the DML statement cannot have any
//	enabled triggers if the statement contains an OUTPUT clause without INTO clause.
//
// The fix is a plain INSERT (no OUTPUT) followed by SCOPE_IDENTITY() to read the new
// row's id back. SCOPE_IDENTITY() - unlike @@IDENTITY - is scoped to the calling batch,
// so it isn't affected by the trigger's own INSERTs into tbl_inv_stock_transactions that
// fire synchronously in between. Both statements are sent as a single Raw/Scan call (one
// batch) rather than two separate tx.Exec/tx.Raw round trips: go-mssqldb sends
// parameterized statements via sp_executesql, and each separate call is its own
// sp_executesql invocation - a stored-procedure scope of its own. A SCOPE_IDENTITY()
// call in a *later*, separate sp_executesql call can't see an identity generated by an
// earlier one, so it comes back NULL even though the insert itself succeeded. Keeping
// the INSERT and the SELECT SCOPE_IDENTITY() in one batch keeps them in the same scope.
func (s *ItemStockService) insertItemStockNoOutput(tx *gorm.DB, body *inventory_models.ItemStocks) error {
	var id uint
	if err := tx.Raw(
		`INSERT INTO tbl_inv_item_stocks (doc_no, item_id, stock_qty, stock_uom, warehouse_id, bin_location, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?);
		 SELECT CAST(SCOPE_IDENTITY() AS INT);`,
		body.DocNo, body.ItemId, body.StockQty, body.StockUom, body.WarehouseId, body.BinLocation, body.IsActive,
	).Scan(&id).Error; err != nil {
		return err
	}
	body.ID = id

	if err := services.InvalidateCache(services.GetKey(body, nil)); err != nil {
		return err
	}
	return services.InvalidateCacheByModel(body)
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

	// Draw the deducted qty from the oldest available lot(s) first (FIFO), so the
	// ledger row below can carry what this specific sale actually cost. RefType/RefId
	// use the same source_type/source_id as the ledger note - see the comment on
	// StockLotConsumption for the one edge case that doesn't cover (multiple lines for
	// the same item+bin under one document).
	cost, err := s.ConsumeLotsFIFO(tx, existing.ItemId, existing.WarehouseId, existing.BinLocation, *body.StockQty, atBody.SourceType, atBody.SourceId)
	if err != nil {
		return nil, fmt.Errorf("failed consuming stock lots: %w", err)
	}

	if err := services.SetStockAuditContext(tx, atBody.SourceType, atBody.SourceId, atBody.Remarks, cost); err != nil {
		return nil, errors.New("failed setting stock audit context")
	}

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

	// Undo whatever ConsumeLotsFIFO drew down for this same (source_type, source_id)
	// when the original deduction happened, oldest-consumed-lot-first in reverse.
	cost, err := s.ReleaseLotsFIFO(tx, atBody.SourceType, atBody.SourceId)
	if err != nil {
		return nil, fmt.Errorf("failed releasing stock lots: %w", err)
	}

	if err := services.SetStockAuditContext(tx, atBody.SourceType, atBody.SourceId, atBody.Remarks, cost); err != nil {
		return nil, errors.New("failed setting stock audit context")
	}

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

// GetStockTransactions returns the trigger-written stock ledger (tbl_inv_stock_transactions)
// - one row per movement, newest first - joined with item/warehouse names, same pattern as
// GetItemStocksList. Optionally filtered to a single item.
func (s *ItemStockService) GetStockTransactions(itemId uint) ([]inventory_models.StockTransactionListView, int, error) {
	var response []inventory_models.StockTransactionListView

	query := `
		SELECT t.id, t.item_id, b.item_code,
		       ISNULL(c.name, '') AS item_name,
		       t.warehouse_id, ISNULL(w.name, '') AS warehouse_name,
		       t.bin_location, t.direction, t.qty_before, t.qty_after, t.qty_change,
		       t.doc_no, t.source_type, t.source_id, t.remarks,
		       t.unit_cost, t.supplier_id, t.supplier, t.purchase_date,
		       t.transaction_at, t.db_user
		FROM tbl_inv_stock_transactions t
		LEFT JOIN tbl_setup_item b ON t.item_id = b.id
		LEFT JOIN tbl_setup_item_name c ON b.item_name_id = c.id
		LEFT JOIN tbl_inv_warehouse_name w ON t.warehouse_id = w.id
		WHERE (? = 0 OR t.item_id = ?)
		ORDER BY t.transaction_at DESC, t.id DESC
	`

	if err := initializers.DB.Raw(query, itemId, itemId).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting stock transactions")
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

	if err := services.SetStockAuditContext(tx, "manual_adjustment", 0, body.Remarks, nil); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed setting stock audit context")
	}

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
