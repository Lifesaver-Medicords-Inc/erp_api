package purchasing_services

import (
	"errors"
	"fmt"

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
		fmt.Println(err)
		fmt.Println("ERR", &PurchaseOrderDetails)
		return errors.New("failed creating purchaseorderdetails")
	}

	quickquotationsat := models.PurchaseOrderDetailsAt{
		RefId:                       PurchaseOrderDetails.ID,
		PurchaseOrderDetailsContent: PurchaseOrderDetails.PurchaseOrderDetailsContent,
		At:                          at,
	}

	if err := services.DbInsert(tx, &quickquotationsat); err != nil {
		return errors.New("failed creating purchaserderdetailsat")
	}

	return nil
}
