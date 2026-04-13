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
// Call this from other services (e.g. receiving report) instead of InsertInventoryStock
// to avoid opening a nested transaction.
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
