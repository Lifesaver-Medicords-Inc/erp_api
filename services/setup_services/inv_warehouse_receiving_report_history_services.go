package setup_services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateReceivingReportHistory(tx *gorm.DB, receivingReportId uint, parentDateReceived string, parentPoId uint, detail models.ReceivingReportDetails2, parentPodId uint, body *ReceivingReportBody2, at models.At) error {
	// Convert OrderedQty and ReceivedQty (strings) to int
	orderedQty, err := strconv.Atoi(detail.OrderedQty)
	if err != nil {
		return errors.New("invalid ordered quantity")
	}

	receivedQty, err := strconv.Atoi(detail.ReceivedQty)
	if err != nil {
		return errors.New("invalid received quantity")
	}

	rejectedQty := 0
	if strings.TrimSpace(detail.RejectedQty) != "" {
		rejectedQty, err = strconv.Atoi(detail.RejectedQty)
		if err != nil {
			return errors.New("invalid rejected quantity")
		}
	}

	// Convert received quantity (string) to int
	receivedQtyInt, err := strconv.Atoi(detail.ReceivedQty)
	if err != nil {
		return errors.New("invalid received quantity in inventory")
	}

	// Convert ordered quantity (string) to int
	orderedQtyInt, err := strconv.Atoi(detail.OrderedQty)
	if err != nil {
		return errors.New("invalid ordered quantity in inventory")
	}

	history := models.ReceivingHistory{
		ReceivingHistoryContent: models.ReceivingHistoryContent{
			PurchaseOrderID:          parentPoId,
			PurchaseOrderDetailsID:   parentPodId,
			ReceivingReportID:        receivingReportId,
			ItemID:                   detail.ItemID,
			ItemCode:                 detail.ItemCode,
			ReceivingReportDetailsID: detail.ID,
			OrderedQty:               strconv.Itoa(orderedQty),
			ReceivedQty:              strconv.Itoa(receivedQty),
			RejectedQty:              strconv.Itoa(rejectedQty),
			DateReceived:             parentDateReceived,
			Uom:                      detail.ReceivedUom,
			BinLocation:              detail.BinLocation,
		},
	}

	inventory := models.InventoryStocks{
		InventoryStocksContent: models.InventoryStocksContent{
			ReceivingReportId:        body.ReceivingReport.ID,
			ReceivingReportDetailsId: detail.ID,
			PurchaseOrderDetailsId:   history.PurchaseOrderDetailsID,
			ReceivingReportDoc:       body.ReceivingReport.DOC,
			PurchaseOrderDoc:         body.ReceivingReport.RefDOC,
			ItemId:                   detail.ItemID,
			ItemCode:                 detail.ItemCode,
			BinLocation:              detail.BinLocation,
			QtyIn:                    uint(receivedQtyInt),
			QtyOut:                   uint(orderedQtyInt),
			Uom:                      detail.ReceivedUom,
			SupplierName:             body.ReceivingReport.SupplierName,
			DateReceived:             body.ReceivingReport.DateReceived,
			WarehouseName:            body.ReceivingReport.WarehouseName,
			WarehouseId:              body.ReceivingReport.WarehouseId,
		},
	}

	// Set current date in MM/dd/yyyy format
	history.TransactionDate = time.Now().Format("01/02/2006")

	if err := CreateInventoryStock(tx, &inventory, at); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &history); err != nil {
		return errors.New("failed creating receiving history")
	}

	historyAt := models.ReceivingHistoryAt{
		RefId:                   history.ID,
		ReceivingHistoryContent: history.ReceivingHistoryContent,
		At:                      at,
	}

	if err := services.DbInsert(tx, &historyAt); err != nil {
		return errors.New("failed creating receiving history at")
	}

	//Invalidate cache
	if err := services.InvalidateCacheByModel(models.PurchaseOrderDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func UpdateReceivingReportHistory(tx *gorm.DB, receivingReportId uint, parentDateReceived string, parentPoId uint, detail models.ReceivingReportDetails2, body *ReceivingReportBody2, at models.At) error {
	// Try to find existing history record for this detail
	var history models.ReceivingHistory
	if err := tx.Where("receiving_report_details_id = ?", detail.ID).First(&history).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("receiving history not found for update")
		}
		return err
	}

	// Convert received quantity (string) to int
	receivedQtyInt, err := strconv.Atoi(detail.ReceivedQty)
	if err != nil {
		return errors.New("invalid received quantity in inventory")
	}

	// Convert ordered quantity (string) to int
	orderedQtyInt, err := strconv.Atoi(detail.OrderedQty)
	if err != nil {
		return errors.New("invalid ordered quantity in inventory")
	}

	// Update fields
	history.ReceivingHistoryContent = models.ReceivingHistoryContent{
		PurchaseOrderID:          parentPoId,
		ReceivingReportID:        receivingReportId,
		ItemCode:                 detail.ItemCode,
		ReceivingReportDetailsID: detail.ID,
		OrderedQty:               detail.OrderedQty,
		ReceivedQty:              detail.ReceivedQty,
		RejectedQty:              detail.RejectedQty,
		DateReceived:             parentDateReceived,
		Uom:                      detail.ReceivedUom,
		BinLocation:              detail.BinLocation,
	}

	inventory := models.InventoryStocks{
		InventoryStocksContent: models.InventoryStocksContent{
			ReceivingReportId:        body.ReceivingReport.ID,
			ReceivingReportDetailsId: detail.ID,
			PurchaseOrderDetailsId:   history.PurchaseOrderDetailsID,
			ReceivingReportDoc:       body.ReceivingReport.DOC,
			PurchaseOrderDoc:         body.ReceivingReport.RefDOC,
			ItemId:                   detail.ItemID,
			ItemCode:                 detail.ItemCode,
			BinLocation:              detail.BinLocation,
			QtyIn:                    uint(receivedQtyInt),
			QtyOut:                   uint(orderedQtyInt),
			Uom:                      detail.ReceivedUom,
			SupplierName:             body.ReceivingReport.SupplierName,
			DateReceived:             body.ReceivingReport.DateReceived,
			WarehouseName:            body.ReceivingReport.WarehouseName,
			WarehouseId:              body.ReceivingReport.WarehouseId,
		},
	}

	if err := UpdateInventoryRRStock(tx, &inventory, at); err != nil {
		return err
	}

	// Set current date in MM/dd/yyyy format
	history.TransactionDate = time.Now().Format("01/02/2006")

	// Save updated history
	if err := services.DbUpdate(tx, &history, map[string]interface{}{"id": history.ID}); err != nil {
		return errors.New("failed updating receiving history")
	}

	// Insert new historyAt (audit trail)
	historyAt := models.ReceivingHistoryAt{
		RefId:                   history.ID,
		ReceivingHistoryContent: history.ReceivingHistoryContent,
		At:                      at,
	}

	if err := services.DbInsert(tx, &historyAt); err != nil {
		return errors.New("failed creating receiving history at")
	}

	//Invalidate cache
	if err := services.InvalidateCacheByModel(models.PurchaseOrderDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func DeleteReceivingReportHistory(tx *gorm.DB, receivingReportID uint, body *DeleteReceivingReportBody2, at models.At) error {
	// Delete ReceivingHistory records where receiving_report_details_id = receivingReportDetailsID
	conditions := map[string]interface{}{
		"receiving_report_id": receivingReportID,
	}

	fmt.Println("delete code: ", conditions)

	if err := services.DbDelete(tx, &models.ReceivingHistory{}, conditions); err != nil {
		return errors.New("failed deleting receiving history")
	}

	if err := DeleteInventoryStock(tx, body.ReceivingReport.ID, at); err != nil {
		return err
	}

	if err := DeleteInvTrackerRR(tx, body.ReceivingReport.ID, at); err != nil {
		return err
	}

	// Insert audit trail record
	receivingReportDetailsAt := models.ReceivingReportDetailsAt2{
		RefId: receivingReportID,
		At:    at,
	}

	if err := services.DbInsert(tx, &receivingReportDetailsAt); err != nil {
		return errors.New("failed creating receiving report details audit trail")
	}

	//Invalidate cache
	if err := services.InvalidateCacheByModel(models.PurchaseOrderDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func DeleteInvTrackerRR(tx *gorm.DB, receivingReportId uint, at models.At) error {
	// Delete all inventory tracker linked to the Receiving Report
	if err := services.DbDelete(tx, &models.InvTracker{}, map[string]interface{}{"rr_id": receivingReportId}); err != nil {
		return errors.New("failed deleting all inventory tracker")
	}

	// Optionally fetch deleted tracker records (Unscoped for audit)
	var deletedInventories []models.InvTracker
	if err := tx.Unscoped().Where("rr_id = ?", receivingReportId).Find(&deletedInventories).Error; err == nil {
		for _, details := range deletedInventories {
			atdataInventory := models.InvTrackerAt{
				RefId:             details.ID,
				InvTrackerContent: details.InvTrackerContent,
				At:                at,
			}
			if err := services.DbInsert(tx, &atdataInventory); err != nil {
				return errors.New("failed creating inventory tracker audit record")
			}
		}
	}

	InvalidateItemCaches()

	return nil
}
