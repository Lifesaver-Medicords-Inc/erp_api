package purchasing_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPurchaseOrderDetails(purchaserderdetails *[]models.PurchaseOrderDetails, conditions map[string]interface{}) error {
	if err := services.DbGet(purchaserderdetails, conditions); err != nil {
		return errors.New("failed getting purchaserderdetails")
	}
	return nil
}

func CreatePurchaseOrderDetails(tx *gorm.DB, basedId uint, PurchaseOrderDetails models.PurchaseOrderDetails, at models.At) error {
	PurchaseOrderDetails.BasedId = basedId

	if err := services.DbInsert(tx, &PurchaseOrderDetails); err != nil {
		return errors.New("failed creating purchaseorderdetails")
	}

	purchaseorderat := models.PurchaseOrderDetailsAt{
		RefId:                       PurchaseOrderDetails.ID,
		PurchaseOrderDetailsContent: PurchaseOrderDetails.PurchaseOrderDetailsContent,
		At:                          at,
	}

	if err := services.DbInsert(tx, &purchaseorderat); err != nil {
		return errors.New("failed creating purchase orderdetailsat")
	}

	if err := services.RecomputeSoItemStatusForCsv(tx, PurchaseOrderDetails.OrderDetailIds); err != nil {
		return errors.New("failed recomputing SO item status")
	}

	return nil
}

func UpdatePurchaseOrderDetails(tx *gorm.DB, basedId uint, purchaserderdetails models.PurchaseOrderDetails, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &purchaserderdetails, conditions); err != nil {
		return errors.New("failed updating purchaseorderdetails")
	}

	purchaseorderdetailsat := models.PurchaseOrderDetailsAt{
		RefId:                       purchaserderdetails.ID,
		PurchaseOrderDetailsContent: purchaserderdetails.PurchaseOrderDetailsContent,
		At:                          at,
	}
	if err := services.DbInsert(tx, &purchaseorderdetailsat); err != nil {
		return errors.New("failed creating purchaseorderdetailsat")
	}

	if err := services.RecomputeSoItemStatusForCsv(tx, purchaserderdetails.OrderDetailIds); err != nil {
		return errors.New("failed recomputing SO item status")
	}

	return nil
}
