package setup_services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateReceivingReportHistory(tx *gorm.DB, receivingReportId uint, parentDateReceived string, parentPoId uint, detail models.ReceivingReportDetails2, parentPodId uint, at models.At) error {
	// Convert OrderedQty and ReceivedQty (strings) to int
	orderedQty, err := strconv.Atoi(detail.OrderedQty)
	if err != nil {
		return errors.New("invalid ordered quantity")
	}

	receivedQty, err := strconv.Atoi(detail.ReceivedQty)
	if err != nil {
		return errors.New("invalid received quantity")
	}

	rejectedQty, err := strconv.Atoi(detail.RejectedQty)
	if err != nil {
		return errors.New("invalid received quantity")
	}

	var remaining = orderedQty - receivedQty

	// Determine completeness
	isComplete := false
	if remaining == 0 {
		isComplete = true
	}

	history := models.ReceivingHistory{
		ReceivingHistoryContent: models.ReceivingHistoryContent{
			PurchaseOrderID:          parentPoId,        // assuming RefId = PO id
			PurchaseOrderDetailsID:   parentPodId,       // assuming RefId = PO id
			ReceivingReportID:        receivingReportId, // assuming RefId = PO id
			ItemCode:                 detail.ItemCode,
			ReceivingReportDetailsID: detail.ID,
			OrderedQty:               strconv.Itoa(remaining),
			ReceivedQty:              strconv.Itoa(receivedQty),
			RejectedQty:              strconv.Itoa(rejectedQty),
			DateReceived:             parentDateReceived, // can use parent DateReceived if needed
			IsComplete:               &isComplete,        // or your logic
		},
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

func UpdateReceivingReportHistory(tx *gorm.DB, receivingReportId uint, parentDateReceived string, parentPoId uint, detail models.ReceivingReportDetails2, at models.At) error {
	// Try to find existing history record for this detail
	var history models.ReceivingHistory
	if err := tx.Where("receiving_report_details_id = ?", detail.ID).First(&history).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("receiving history not found for update")
		}
		return err
	}

	// Convert OrderedQty and ReceivedQty (strings) to int
	orderedQty, err := strconv.Atoi(history.OrderedQty)
	if err != nil {
		return errors.New("invalid ordered quantity")
	}

	receivedQty, err := strconv.Atoi(history.ReceivedQty)
	if err != nil {
		return errors.New("invalid received quantity")
	}

	receivedDetailQty, err := strconv.Atoi(detail.ReceivedQty)
	if err != nil {
		return errors.New("invalid ordered quantity")
	}

	// Default completeness
	isComplete := false

	if receivedDetailQty < receivedQty {
		diff := receivedQty - receivedDetailQty
		orderedQty += diff
	} else if receivedDetailQty > receivedQty {
		diff := receivedDetailQty - receivedQty
		orderedQty -= diff
	}

	// Final completeness check
	if orderedQty == 0 {
		isComplete = true
	}

	// Update fields
	history.ReceivingHistoryContent = models.ReceivingHistoryContent{
		PurchaseOrderID:          parentPoId,
		ReceivingReportID:        receivingReportId,
		ItemCode:                 detail.ItemCode,
		ReceivingReportDetailsID: detail.ID,
		OrderedQty:               strconv.Itoa(orderedQty),
		ReceivedQty:              detail.ReceivedQty,
		RejectedQty:              detail.RejectedQty,
		DateReceived:             parentDateReceived,
		IsComplete:               &isComplete,
	}

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

func DeleteReceivingReportHistory(tx *gorm.DB, receivingReportID uint, at models.At) error {
	// Delete ReceivingHistory records where receiving_report_details_id = receivingReportDetailsID
	conditions := map[string]interface{}{
		"receiving_report_id": receivingReportID,
	}

	fmt.Println("delete code: ", conditions)

	if err := services.DbDelete(tx, &models.ReceivingHistory{}, conditions); err != nil {
		return errors.New("failed deleting receiving history")
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
