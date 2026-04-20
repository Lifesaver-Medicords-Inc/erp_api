package item_request_services

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

type ItemRequestService struct {
	stockService *item_stock_services.ItemStockService
}

func NewItemRequestService() *ItemRequestService {
	return &ItemRequestService{
		stockService: item_stock_services.NewItemStockService(),
	}
}

func (s *ItemRequestService) GetItemRequest(conditions map[string]interface{}) (interface{}, int, error) {
	var response inventory_models.ItemRequestGet

	if err := services.DbGet(&response.ItemRequest, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request")
	}

	if err := services.DbGet(&response.ItemRequestDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request details")
	}

	if err := services.DbGet(&response.ItemRequestLocations, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request locations")
	}

	return response, fiber.StatusOK, nil
}

func (s *ItemRequestService) GetUserList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.UserListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting user list")
	}

	return response, fiber.StatusOK, nil
}

func (s *ItemRequestService) GetAllItems(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.AllItemView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item list")
	}

	return response, fiber.StatusOK, nil
}

func (s *ItemRequestService) GetItemReqSODoc(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.SalesOrderItemReqDocView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order doc")
	}

	return response, fiber.StatusOK, nil
}

func (s *ItemRequestService) GetItemReqSO(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.SalesOrderItemReqDetailsView

	if err := services.DbRaw(&response, "sp_GetSalesOrderDetailsItemReq", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order details")
	}

	return response, fiber.StatusOK, nil
}

func (s *ItemRequestService) CreateItemRequest(body *inventory_models.ItemRequestBody, at models.At) (*inventory_models.ItemRequestBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(inventory_models.ItemRequest), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.ItemRequest.DocNo = nextDocNo

	if err := services.DbInsert(tx, &body.ItemRequest); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request")
	}

	if err := s.CreateItemRequestDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ItemRequestAt{
		RefId:              body.ItemRequest.ID,
		ItemRequestContent: body.ItemRequest.ItemRequestContent,
		At:                 at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateItemRequestCaches()

	return body, fiber.StatusOK, nil
}

func (s *ItemRequestService) CreateItemRequestDetails(tx *gorm.DB, body *inventory_models.ItemRequestBody, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.ItemRequestId = body.ItemRequest.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating item request details")
		}

		atdataDetail := inventory_models.ItemRequestDetailsAt{
			RefId:                     detail.ID,
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating item request details at")
		}
	}
	return nil
}

// CreateItemRequestLocations upserts locations for a single detail line.
// Called only from UpdateItemRequestDetails — never standalone.
func (s *ItemRequestService) CreateItemRequestLocations(tx *gorm.DB, detail *inventory_models.ItemRequestDetails, locations []inventory_models.ItemRequestLocations, at models.At) error {
	for i := range locations {
		loc := &locations[i]

		// Bind to parent keys — enforced here, not caller's responsibility
		loc.ItemRequestId = detail.ItemRequestId
		loc.ItemRequestDetailsId = detail.ID

		if loc.ID == 0 {
			if err := services.DbInsert(tx, loc); err != nil {
				return fmt.Errorf("failed creating item request location for detail %d: %w", detail.ID, err)
			}
		} else {
			if err := services.DbUpdate(tx, loc, map[string]interface{}{"id": loc.ID}); err != nil {
				return fmt.Errorf("failed updating item request location %d: %w", loc.ID, err)
			}
		}

		atdata := inventory_models.ItemRequestLocationsAt{
			RefId:                       loc.ID,
			ItemRequestLocationsContent: loc.ItemRequestLocationsContent,
			At:                          at,
		}
		if err := services.DbInsert(tx, &atdata); err != nil {
			return fmt.Errorf("failed creating item request location audit for detail %d: %w", detail.ID, err)
		}

		// Build the stock body from the detail fields and call UpsertStockWithTx directly
		stockBody := &inventory_models.ItemStocks{
			ID: loc.BinId,
			ItemStocksContent: inventory_models.ItemStocksContent{
				StockQty: &loc.SelectedQty,
			},
		}

		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   detail.ID,
			SourceType: "receiving_report",
		}

		if _, err := s.stockService.DeductStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("failed deduct inventory stock for item %d: %w", detail.ItemId, err)
		}
	}

	return nil
}

func (s *ItemRequestService) UpdateItemRequest(body *inventory_models.ItemRequestBody, conditions map[string]interface{}, at models.At) (*inventory_models.ItemRequestBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body.ItemRequest, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item request")
	}

	if err := s.UpdateItemRequestDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ItemRequestAt{
		RefId:              body.ItemRequest.ID,
		ItemRequestContent: body.ItemRequest.ItemRequestContent,
		At:                 at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item request at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateItemRequestCaches()

	return body, fiber.StatusOK, nil
}

func (s *ItemRequestService) UpdateItemRequestDetails(tx *gorm.DB, body *inventory_models.ItemRequestBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.ItemRequestId = body.ItemRequest.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating item request details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating item request details")
			}
		}

		atdataDetail := inventory_models.ItemRequestDetailsAt{
			RefId:                     detail.ID,
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating item request details at")
		}

		// Filter locations belonging to this detail line only
		var detailLocations []inventory_models.ItemRequestLocations
		for _, loc := range body.ItemRequestLocations {
			if loc.ItemRequestDetailsId == detail.ID {
				detailLocations = append(detailLocations, loc)
			}
		}

		if len(detailLocations) > 0 {
			if err := s.CreateItemRequestLocations(tx, detail, detailLocations, at); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ItemRequestService) DeleteItemRequest(body *inventory_models.ItemRequestBody, at models.At) (*inventory_models.ItemRequestBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body.ItemRequest, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item request")
	}

	if err := s.DeleteItemRequestDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ItemRequestAt{RefId: body.ItemRequest.ID, ItemRequestContent: body.ItemRequest.ItemRequestContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateItemRequestCaches()

	return body, fiber.StatusOK, nil
}

func (s *ItemRequestService) DeleteItemRequestDetails(tx *gorm.DB, body *inventory_models.ItemRequestBody, at models.At) error {
	itemRequestId := body.ItemRequest.ID

	// --- 1. Audit then delete locations ---
	var deletedLocations []inventory_models.ItemRequestLocations
	if err := tx.Unscoped().
		Where("item_request_id = ?", itemRequestId).
		Find(&deletedLocations).Error; err == nil {
		for _, loc := range deletedLocations {
			atLoc := inventory_models.ItemRequestLocationsAt{
				RefId:                       loc.ID,
				ItemRequestLocationsContent: loc.ItemRequestLocationsContent,
				At:                          at,
			}
			if err := services.DbInsert(tx, &atLoc); err != nil {
				return errors.New("failed creating item request location audit record")
			}

			// Restore the deducted qty back to its bin
			stockBody := &inventory_models.ItemStocks{
				ID: loc.BinId,
				ItemStocksContent: inventory_models.ItemStocksContent{
					StockQty: &loc.SelectedQty,
				},
			}
			stockAtBody := &inventory_models.ItemStocksAt{
				SourceId:   loc.ItemRequestDetailsId,
				SourceType: "item_request_deletion",
			}
			if _, err := s.stockService.RestoreStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
				return fmt.Errorf("failed restoring stock for location %d (bin %d): %w", loc.ID, loc.BinId, err)
			}
		}
	}

	if err := services.DbDelete(tx, &inventory_models.ItemRequestLocations{},
		map[string]interface{}{"item_request_id": itemRequestId}); err != nil {
		return errors.New("failed deleting item request locations")
	}

	// --- 2. Audit then delete details ---
	if err := services.DbDelete(tx, &inventory_models.ItemRequestDetails{},
		map[string]interface{}{"item_request_id": itemRequestId}); err != nil {
		return errors.New("failed deleting all item request details")
	}

	var deletedDetails []inventory_models.ItemRequestDetails
	if err := tx.Unscoped().
		Where("item_request_id = ?", itemRequestId).
		Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := inventory_models.ItemRequestDetailsAt{
				RefId:                     detail.ID,
				ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
				At:                        at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating item request details audit record")
			}

		}
	}

	return nil
}

func invalidateItemRequestCaches() {
	setup_services.InvalidateItemCaches()

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDocView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.ItemRequestDetailsGet{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}
}
