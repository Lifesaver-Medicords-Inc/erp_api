package pick_activity_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type PickActivityService struct {
	stockService *item_stock_services.ItemStockService
}

func NewPickActivityService() *PickActivityService {
	return &PickActivityService{
		stockService: item_stock_services.NewItemStockService(),
	}
}

func (s *PickActivityService) GetPickActivity(conditions map[string]interface{}) (interface{}, int, error) {
	var response inventory_models.PickActivityGet

	if err := services.DbGet(&response.PickActivity, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity")
	}

	if err := services.DbGet(&response.PickActivityDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity details")
	}

	if err := services.DbGet(&response.PickActivityLocations, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity locations")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetWarehousePickAct(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting warehouse")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetWarehouseAreaPickAct(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingAreaView

	if err := services.DbRaw(&response, "sp_GetWarehouseAreaReceiving", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting warehouse area data")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetPickActSODoc(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.SalesOrderPickActDocView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order doc")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetPickActSO(conditions map[string]interface{}) (interface{}, int, error) {
	var poParent []inventory_models.SalesOrderPickActView

	// Get Sales Order (Parent)
	if err := services.DbRaw(&poParent, "sp_GetSalesOrderPickAct", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order data")
	}

	if len(poParent) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no sales order found")
	}

	var allChildren []inventory_models.SalesOrderPickActDetailsView

	for _, so := range poParent {
		var poChild []inventory_models.SalesOrderPickActDetailsView

		childConditions := map[string]interface{}{
			"SalesId": so.SalesOrderId,
		}

		if err := services.DbRaw(&poChild, "sp_GetSalesOrderDetailsPickAct", childConditions); err != nil {
			return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order details data")
		}

		// Append children to one slice
		allChildren = append(allChildren, poChild...)
	}

	// Response structure
	response := struct {
		SalesOrderView        []inventory_models.SalesOrderPickActView        `json:"sales_order_view"`
		SalesOrderDetailsView []inventory_models.SalesOrderPickActDetailsView `json:"sales_order_details_view"`
	}{
		SalesOrderView:        poParent,
		SalesOrderDetailsView: allChildren,
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) CreatePickActivity(body *inventory_models.PickActivityBody, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(inventory_models.PickActivity), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.PickActivity.DocNo = nextDocNo

	if err := services.DbInsert(tx, &body.PickActivity); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity")
	}

	if err := s.CreatePickActivityDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) CreatePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PickActivityId = body.PickActivity.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating pick activity details")
		}

		atdataDetail := inventory_models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}

		if err := services.RecomputeSoItemStatus(tx, detail.SalesOrderDetailsId); err != nil {
			return errors.New("failed recomputing SO item status")
		}
	}
	return nil
}

// CreatePickActivityLocations upserts locations for a single detail line.
// Called only from UpdatePickActivityDetails — never standalone.
func (s *PickActivityService) CreatePickActivityLocations(tx *gorm.DB, detail *inventory_models.PickActivityDetails, locations []inventory_models.PickActivityLocations, at models.At) error {
	// detail.PickActivityId IS the Pick Activity header's own primary key, so it can be
	// used directly as the ledger's source_id — no extra lookup needed for that part.
	// DocNo isn't on the detail row though, so fetch it once here for a readable note.
	var paDocNo int
	tx.Model(&inventory_models.PickActivity{}).Select("doc_no").Where("id = ?", detail.PickActivityId).Scan(&paDocNo)
	paRemarks := fmt.Sprintf("Pick Activity #%d", paDocNo)

	for i := range locations {
		loc := &locations[i]

		// Bind to parent keys — enforced here, not caller's responsibility
		loc.PickActivityId = detail.PickActivityId
		loc.PickActivityDetailsId = detail.ID

		if loc.ID == 0 {
			if err := services.DbInsert(tx, loc); err != nil {
				return fmt.Errorf("failed creating pick activity location for detail %d: %w", detail.ID, err)
			}
		} else {
			if err := services.DbUpdate(tx, loc, map[string]interface{}{"id": loc.ID}); err != nil {
				return fmt.Errorf("failed updating pick activity location %d: %w", loc.ID, err)
			}
		}

		atdata := inventory_models.PickActivityLocationsAt{
			RefId:                        loc.ID,
			PickActivityLocationsContent: loc.PickActivityLocationsContent,
			At:                           at,
		}
		if err := services.DbInsert(tx, &atdata); err != nil {
			return fmt.Errorf("failed creating pick activity location audit for detail %d: %w", detail.ID, err)
		}

		// Build the stock body from the detail fields and call UpsertStockWithTx directly
		stockBody := &inventory_models.ItemStocks{
			ID: loc.BinId,
			ItemStocksContent: inventory_models.ItemStocksContent{
				StockQty: &loc.SelectedQty,
			},
		}

		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   detail.PickActivityId,
			SourceType: "pick_activity",
			Remarks:    paRemarks,
		}

		if _, err := s.stockService.DeductStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("failed deduct inventory stock for item %d: %w", detail.ItemId, err)
		}
	}

	return nil
}

func (s *PickActivityService) UpdatePickActivity(body *inventory_models.PickActivityBody, conditions map[string]interface{}, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body.PickActivity, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity")
	}

	if err := s.UpdatePickActivityDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) UpdatePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PickActivityId = body.PickActivity.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating pick activity details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating pick activity details")
			}
		}

		atdataDetail := inventory_models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}

		if detail.ActualQty > 0 {
			stockBody := &inventory_models.ItemStocks{
				ItemStocksContent: inventory_models.ItemStocksContent{
					ItemId:      detail.ItemId,
					StockQty:    &detail.ActualQty,
					StockUom:    detail.ActualUom,
					WarehouseId: detail.WarehouseId,
					BinLocation: detail.BinLocation,
				},
			}

			stockAtBody := &inventory_models.ItemStocksAt{
				SourceId:   body.PickActivity.ID,
				SourceType: "pick_activity",
				Remarks:    fmt.Sprintf("Pick Activity #%d", body.PickActivity.DocNo),
			}

			if _, err := s.stockService.UpsertStockWithTx(tx, stockBody, stockAtBody, at, nil); err != nil {
				return fmt.Errorf("failed upserting inventory stock for item %d: %w", detail.ItemId, err)
			}
		}

		// Filter locations belonging to this detail line only
		var detailLocations []inventory_models.PickActivityLocations
		for _, loc := range body.PickActivityLocations {
			if loc.PickActivityDetailsId == detail.ID {
				detailLocations = append(detailLocations, loc)
			}
		}

		if len(detailLocations) > 0 {
			if err := s.CreatePickActivityLocations(tx, detail, detailLocations, at); err != nil {
				return err
			}
		}

		if err := services.RecomputeSoItemStatus(tx, detail.SalesOrderDetailsId); err != nil {
			return errors.New("failed recomputing SO item status")
		}
	}

	return nil
}

func (s *PickActivityService) DeletePickActivity(body *inventory_models.PickActivityBody, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body.PickActivity, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting pick activity")
	}

	if err := s.DeletePickActivityDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{RefId: body.PickActivity.ID, PickActivityContent: body.PickActivity.PickActivityContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) DeletePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, at models.At) error {
	pickActivityId := body.PickActivity.ID

	// Delete requests typically only carry the Pick Activity's ID, not its DocNo, so
	// fetch the doc number once here rather than showing "#0" in the ledger notes below.
	var paDocNo int
	tx.Model(&inventory_models.PickActivity{}).Select("doc_no").Where("id = ?", pickActivityId).Scan(&paDocNo)
	paDeletionRemarks := fmt.Sprintf("Pick Activity #%d (deleted)", paDocNo)

	// --- 1. Fetch locations BEFORE deleting (need data for audit + stock restore) ---
	var deletedLocations []inventory_models.PickActivityLocations
	if err := tx.Unscoped().
		Where("pick_activity_id = ?", pickActivityId).
		Find(&deletedLocations).Error; err != nil {
		return errors.New("failed fetching pick activity locations for deletion")
	}

	for _, loc := range deletedLocations {
		atLoc := inventory_models.PickActivityLocationsAt{
			RefId:                        loc.ID,
			PickActivityLocationsContent: loc.PickActivityLocationsContent,
			At:                           at,
		}
		if err := services.DbInsert(tx, &atLoc); err != nil {
			return errors.New("failed creating pick activity location audit record")
		}

		stockBody := &inventory_models.ItemStocks{
			ID: loc.BinId,
			ItemStocksContent: inventory_models.ItemStocksContent{
				StockQty: &loc.SelectedQty,
			},
		}
		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   pickActivityId,
			SourceType: "pick_activity_deletion",
			Remarks:    paDeletionRemarks,
		}
		if _, err := s.stockService.RestoreStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("failed restoring stock for location %d (bin %d): %w", loc.ID, loc.BinId, err)
		}
	}

	// --- 2. Delete locations (before details, FK order) ---
	if err := services.DbDelete(tx, &inventory_models.PickActivityLocations{},
		map[string]interface{}{"pick_activity_id": pickActivityId}); err != nil {
		return errors.New("failed deleting pick activity locations")
	}

	// --- 3. Fetch details BEFORE deleting ---
	var deletedDetails []inventory_models.PickActivityDetails
	if err := tx.Unscoped().
		Where("pick_activity_id = ?", pickActivityId).
		Find(&deletedDetails).Error; err != nil {
		return errors.New("failed fetching pick activity details for deletion")
	}

	for _, detail := range deletedDetails {
		// Reverse the stock that was added when this detail was created
		var stock inventory_models.ItemStocks
		err := tx.Where(
			"item_id = ? AND warehouse_id = ? AND bin_location = ?",
			detail.ItemId, detail.WarehouseId, detail.BinLocation,
		).First(&stock).Error

		if err == nil {
			// Known gap, not fixed here: doesn't touch tbl_inv_stock_lots (this Upsert
			// created a lot via UpsertStockWithTx when ActualQty > 0 - see
			// UpdatePickActivityDetails above). Same class of gap as
			// receiving_report_service.go's equivalent block.
			*stock.StockQty -= detail.ActualQty
			s.stockService.SetActiveStatus(&stock) // flips IsActive to false if qty hits zero

			if err := services.SetStockAuditContext(tx, "pick_activity_delete", pickActivityId, paDeletionRemarks, nil); err != nil {
				return errors.New("failed setting stock audit context")
			}

			if err := services.DbUpdate(tx, &stock, map[string]interface{}{"id": stock.ID}); err != nil {
				return errors.New("failed reversing inventory stock for deleted detail")
			}

			// Audit the reversal
			atStock := inventory_models.ItemStocksAt{
				RefId:             stock.ID,
				SourceId:          pickActivityId,
				SourceType:        "pick_activity_delete",
				Remarks:           paDeletionRemarks,
				ItemStocksContent: stock.ItemStocksContent,
				At:                at,
			}
			if err := services.DbInsert(tx, &atStock); err != nil {
				return errors.New("failed creating stock reversal audit record")
			}
		}

		atdataDetail := inventory_models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details audit record")
		}
	}

	// --- 4. Delete details ---
	if err := services.DbDelete(tx, &inventory_models.PickActivityDetails{},
		map[string]interface{}{"pick_activity_id": pickActivityId}); err != nil {
		return errors.New("failed deleting pick activity details")
	}

	return nil
}

func invalidatePickActivityCaches() {
	setup_services.InvalidateItemCaches()

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDocView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.PickActivityDetailsGet{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.ItemLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}
}
