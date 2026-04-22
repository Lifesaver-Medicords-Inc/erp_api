package receiving_report_services

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

type ReceivingReportService struct {
	stockService *item_stock_services.ItemStockService
}

func NewReceivingReportService() *ReceivingReportService {
	return &ReceivingReportService{
		stockService: item_stock_services.NewItemStockService(),
	}
}

func (s *ReceivingReportService) GetReceivingReport(conditions map[string]interface{}) (interface{}, int, error) {
	var response inventory_models.ReceivingReportGet

	if err := services.DbGet(&response.ReceivingReport, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting receiving report")
	}

	if err := services.DbGet(&response.ReceivingReportDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting receiving report details")
	}

	return response, fiber.StatusOK, nil
}

func (s *ReceivingReportService) GetWarehouseReceiving(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting warehouse")
	}

	return response, fiber.StatusOK, nil
}

func (s *ReceivingReportService) GetWarehouseAreaReceiving(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingAreaView

	if err := services.DbRaw(&response, "sp_GetWarehouseAreaReceiving", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting warehouse area data")
	}

	return response, fiber.StatusOK, nil
}

func (s *ReceivingReportService) GetReceivingPODoc(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.PurchaseOrderDocView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting purchase order doc")
	}

	return response, fiber.StatusOK, nil
}

func (s *ReceivingReportService) GetReceivingPO(conditions map[string]interface{}) (interface{}, int, error) {
	var poParent []inventory_models.PurchaseOrderReceivingView

	if err := services.DbRaw(&poParent, "sp_GetPurchaseOrderReceiving", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting purchase order data")
	}

	if len(poParent) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no purchase order found")
	}

	var allChildren []inventory_models.PurchaseOrderReceivingDetailsView

	for _, po := range poParent {
		var poChild []inventory_models.PurchaseOrderReceivingDetailsView

		childConditions := map[string]interface{}{
			"PurchaseId": po.PurchaseOrderId,
		}

		if err := services.DbRaw(&poChild, "sp_GetPurchaseOrderDetailsReceiving", childConditions); err != nil {
			return nil, fiber.StatusInternalServerError, errors.New("failed getting purchase order details data")
		}

		allChildren = append(allChildren, poChild...)
	}

	response := struct {
		PurchaseOrderView        []inventory_models.PurchaseOrderReceivingView        `json:"purchase_order_view"`
		PurchaseOrderDetailsView []inventory_models.PurchaseOrderReceivingDetailsView `json:"purchase_order_details_view"`
	}{
		PurchaseOrderView:        poParent,
		PurchaseOrderDetailsView: allChildren,
	}

	return response, fiber.StatusOK, nil
}

func (s *ReceivingReportService) CreateReceivingReport(body *inventory_models.ReceivingReportBody, at models.At) (*inventory_models.ReceivingReportBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(inventory_models.ReceivingReport), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.ReceivingReport.DocNo = nextDocNo

	if err := services.DbInsert(tx, &body.ReceivingReport); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report")
	}

	if err := s.CreateReceivingReportDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ReceivingReportAt{
		RefId:                  body.ReceivingReport.ID,
		ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateReceivingCaches()

	return body, fiber.StatusOK, nil
}

func (s *ReceivingReportService) CreateReceivingReportDetails(tx *gorm.DB, body *inventory_models.ReceivingReportBody, at models.At) error {
	for i := range body.ReceivingReportDetails {
		detail := &body.ReceivingReportDetails[i]
		detail.ReceivingReportId = body.ReceivingReport.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating receiving report details")
		}

		atdataDetail := inventory_models.ReceivingReportDetailsAt{
			RefId:                         detail.ID,
			ReceivingReportDetailsContent: detail.ReceivingReportDetailsContent,
			At:                            at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating receiving report details at")
		}

		// Build the stock body from the detail fields and call UpsertStockWithTx directly
		stockBody := &inventory_models.ItemStocks{
			ItemStocksContent: inventory_models.ItemStocksContent{
				ItemId:      detail.ItemID,
				StockQty:    &detail.ReceivedQty,
				StockUom:    detail.ReceivedUom,
				WarehouseId: detail.WarehouseId,
				BinLocation: detail.BinLocation,
			},
		}

		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   detail.ID,
			SourceType: "receiving_report",
		}

		if _, err := s.stockService.UpsertStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("failed upserting inventory stock for item %d: %w", detail.ItemID, err)
		}
	}
	return nil
}

func (s *ReceivingReportService) UpdateReceivingReport(body *inventory_models.ReceivingReportBody, conditions map[string]interface{}, at models.At) (*inventory_models.ReceivingReportBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating receiving report")
	}

	if err := s.UpdateReceivingReportDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ReceivingReportAt{
		RefId:                  body.ReceivingReport.ID,
		ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating receiving report at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateReceivingCaches()

	return body, fiber.StatusOK, nil
}

func (s *ReceivingReportService) UpdateReceivingReportDetails(tx *gorm.DB, body *inventory_models.ReceivingReportBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.ReceivingReportDetails {
		detail := &body.ReceivingReportDetails[i]
		detail.ReceivingReportId = body.ReceivingReport.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating receiving report details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating receiving report details")
			}
		}

		atdataDetail := inventory_models.ReceivingReportDetailsAt{
			RefId:                         detail.ID,
			ReceivingReportDetailsContent: detail.ReceivingReportDetailsContent,
			At:                            at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating receiving report details at")
		}
	}

	return nil
}

func (s *ReceivingReportService) DeleteReceivingReport(body *inventory_models.ReceivingReportBody, at models.At) (*inventory_models.ReceivingReportBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body.ReceivingReport, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting receiving report")
	}

	if err := s.DeleteReceivingReportDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.ReceivingReportAt{RefId: body.ReceivingReport.ID, ReceivingReportContent: body.ReceivingReport.ReceivingReportContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidateReceivingCaches()

	return body, fiber.StatusOK, nil
}

func (s *ReceivingReportService) DeleteReceivingReportDetails(tx *gorm.DB, body *inventory_models.ReceivingReportBody, at models.At) error {
	var deletedDetails []inventory_models.ReceivingReportDetails
	if err := tx.Where("receiving_report_id = ?", body.ReceivingReport.ID).Find(&deletedDetails).Error; err != nil {
		return errors.New("failed fetching receiving report details for deletion")
	}

	for _, detail := range deletedDetails {
		// Reverse the stock that was added when this detail was created
		var stock inventory_models.ItemStocks
		err := tx.Where(
			"item_id = ? AND warehouse_id = ? AND bin_location = ?",
			detail.ItemID, detail.WarehouseId, detail.BinLocation,
		).First(&stock).Error

		if err == nil {
			*stock.StockQty -= detail.ReceivedQty
			s.stockService.SetActiveStatus(&stock) // flips IsActive to false if qty hits zero

			if err := services.DbUpdate(tx, &stock, map[string]interface{}{"id": stock.ID}); err != nil {
				return errors.New("failed reversing inventory stock for deleted detail")
			}

			// Audit the reversal
			atStock := inventory_models.ItemStocksAt{
				RefId:             stock.ID,
				SourceId:          detail.ID,
				SourceType:        "receiving_report_delete",
				ItemStocksContent: stock.ItemStocksContent,
				At:                at,
			}
			if err := services.DbInsert(tx, &atStock); err != nil {
				return errors.New("failed creating stock reversal audit record")
			}
		}

		// Audit the detail deletion
		atdataDetail := inventory_models.ReceivingReportDetailsAt{
			RefId:                         detail.ID,
			ReceivingReportDetailsContent: detail.ReceivingReportDetailsContent,
			At:                            at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating receiving report details audit record")
		}
	}

	// Delete the detail rows after processing
	if err := services.DbDelete(tx, &inventory_models.ReceivingReportDetails{}, map[string]interface{}{"receiving_report_id": body.ReceivingReport.ID}); err != nil {
		return errors.New("failed deleting all receiving report details")
	}

	return nil
}

func invalidateReceivingCaches() {
	setup_services.InvalidateItemCaches()

	if err := services.InvalidateCacheByModel(inventory_models.WarehouseReceivingAreaView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.PurchaseOrderReceivingView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.PurchaseOrderReceivingDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.ReceivingReportDetailsGet{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.ItemLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}
}
