package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BoqDetails(tx *gorm.DB, parentId uint, child models.ItemBoqDetails, at models.At) error {

	child.ItemBoqDetailsContent.ItemBoqID = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to create bom details")
	}

	childAt := models.ItemBoqDetailsAt{
		RefId:                 child.ID,
		ItemBoqDetailsContent: child.ItemBoqDetailsContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bom_details_at")
	}

	return nil
}

func UpdateBoqChild(tx *gorm.DB, ItemBoqDetails models.ItemBoqDetails, at models.At, boqId uint) error {
	if ItemBoqDetails.ID == 0 {
		ItemBoqDetails.ItemBoqDetailsContent.ItemBoqID = boqId
		if err := services.DbInsert(tx, &ItemBoqDetails); err != nil {
			return errors.New("failed to create boq details")
		}
	} else {
		conditions := map[string]interface{}{
			"id": ItemBoqDetails.ID,
		}

		if err := services.DbUpdate(tx, &ItemBoqDetails, conditions); err != nil {
			return errors.New("failed updating child")
		}
	}

	ItemBoqDetailsAt := models.ItemBoqDetailsAt{
		RefId:                 ItemBoqDetails.ID,
		ItemBoqDetailsContent: ItemBoqDetails.ItemBoqDetailsContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &ItemBoqDetailsAt); err != nil {
		return errors.New("failed creating ItemBoqDetailsAt")
	}

	return nil
}
