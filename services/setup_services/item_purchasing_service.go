package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateItemPurchasing(tx *gorm.DB, basedId uint, itemPurchasing models.ItemPurchasing, at models.At) error {
	content := models.ItemPurchasingContent{
		BasedId:          basedId,
		SupplierNameId:   itemPurchasing.SupplierNameId,
		PaymentTermsId:   itemPurchasing.PaymentTermsId,
		Price:            itemPurchasing.Price,
		SupplierTypeName: itemPurchasing.SupplierTypeName,
		ValidityPeriod:   itemPurchasing.ValidityPeriod,
	}

	itempurchasing := models.ItemPurchasing{ItemPurchasingContent: content}

	if err := services.DbInsert(tx, &itempurchasing); err != nil {
		return errors.New("failed creating item purchasing")
	}

	itempurchasingat := models.ItemPurchasingAt{
		RefId:                 itemPurchasing.ID,
		ItemPurchasingContent: content,
		At:                    at,
	}

	if err := services.DbInsert(tx, &itempurchasingat); err != nil {
		return errors.New("failed creating itempurchasingat")
	}

	return nil
}

func UpdateItemPurchasing(tx *gorm.DB, itemPurchasing models.ItemPurchasing, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &itemPurchasing, conditions); err != nil {
		return errors.New("failed updating additional specs")
	}

	additionalspecat := models.ItemPurchasingAt{
		RefId:                 itemPurchasing.ID,
		ItemPurchasingContent: itemPurchasing.ItemPurchasingContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &additionalspecat); err != nil {
		return errors.New("failed creating additional specs at")
	}

	return nil
}
