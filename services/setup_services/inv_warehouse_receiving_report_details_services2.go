package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetReceivingReportDetails2(conditions map[string]interface{}) ([]models.ReceivingReportDetails2, int, error) {
	var receivingReportDetails []models.ReceivingReportDetails2

	if err := services.DbGet(&receivingReportDetails, conditions); err != nil {
		return receivingReportDetails, fiber.StatusInternalServerError, errors.New("failed getting warehouse address")
	}

	return receivingReportDetails, 0, nil
}

func CreateReceivingReportDetails2(tx *gorm.DB, parentId uint, parentDate string, parentPoId uint, child models.ReceivingReportDetails2, at models.At) error {
	content := models.ReceivingReportDetailsContent2{
		ReceivingReportId:  parentId,
		PodId:              child.PodId,
		ItemCode:           child.ItemCode,
		ItemDescription:    child.ItemDescription,
		OrderedQty:         child.OrderedQty,
		OrderedUom:         child.OrderedUom,
		ReceivedQty:        child.ReceivedQty,
		ReceivedUom:        child.ReceivedUom,
		RejectedQty:        child.RejectedQty,
		RejectedUom:        child.RejectedUom,
		SerialNumber:       child.SerialNumber,
		BinLocation:        child.BinLocation,
		ReasonForRejection: child.ReasonForRejection,
		RefId:              child.RefId,
	}

	receivingReportDetails := models.ReceivingReportDetails2{
		ReceivingReportDetailsContent2: content,
	}
	if err := services.DbInsert(tx, &receivingReportDetails); err != nil {
		return errors.New("failed creating warehouse area")
	}

	ReceivingReportDetailsAt := models.ReceivingReportDetailsAt2{
		RefId:                          receivingReportDetails.ID,
		ReceivingReportDetailsContent2: content,
		At:                             at,
	}
	if err := services.DbInsert(tx, &ReceivingReportDetailsAt); err != nil {
		return errors.New("failed creating warehouse area at")
	}

	if err := CreateReceivingReportHistory(tx, parentId, parentDate, parentPoId, receivingReportDetails, child.PodId, at); err != nil {
		return err
	}

	return nil
}

func UpdateReceivingReportDetails2(tx *gorm.DB, ReceivingReportDetails models.ReceivingReportDetails2, at models.At, conditions map[string]interface{}, parentId uint, parentDate string, parentPoId uint) error {

	if err := services.DbUpdate(tx, &ReceivingReportDetails, conditions); err != nil {
		return errors.New("failed updating receiving report details")
	}

	ReceivingReportDetailsAt := models.ReceivingReportDetailsAt2{
		RefId:                          ReceivingReportDetails.ID,
		ReceivingReportDetailsContent2: ReceivingReportDetails.ReceivingReportDetailsContent2,
		At:                             at,
	}

	if err := services.DbInsert(tx, &ReceivingReportDetailsAt); err != nil {
		return errors.New("failed creating receiving report details at")
	}

	// Add history record after update
	if err := UpdateReceivingReportHistory(tx, parentId, parentDate, parentPoId, ReceivingReportDetails, at); err != nil {
		return err
	}

	return nil
}

func DeleteReceivingReportDetails2(tx *gorm.DB, ReceivingReportDetails models.ReceivingReportDetails2, at models.At, conditions map[string]interface{}) error {

	if err := services.DbDelete(tx, &ReceivingReportDetails, conditions); err != nil {
		return errors.New("failed deleting receiving report details")
	}

	receivingreportdetailsat := models.ReceivingReportDetailsAt2{
		RefId:                          ReceivingReportDetails.ID,
		ReceivingReportDetailsContent2: ReceivingReportDetails.ReceivingReportDetailsContent2,
		At:                             at,
	}
	if err := services.DbInsert(tx, &receivingreportdetailsat); err != nil {
		return errors.New("failed creating receiving report detailsat at")
	}

	return nil
}
