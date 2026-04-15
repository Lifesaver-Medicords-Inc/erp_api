package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BomDetails(tx *gorm.DB, parentId uint, child models.SetupItemBomDetails, at models.At) error {
	child.ItemBomID = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to create bom details")
	}

	childAt := models.SetupItemBomDetailsAt{
		RefId:                      child.ID,
		SetupItemBomDetailsContent: child.SetupItemBomDetailsContent,
		At:                         at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bom_details_at")
	}

	return nil
}

func UpdateChild(tx *gorm.DB, setupItemBomDetails models.SetupItemBomDetails, at models.At, bomId uint) error {
	if setupItemBomDetails.ID == 0 {
		setupItemBomDetails.ItemBomID = bomId
		if err := services.DbInsert(tx, &setupItemBomDetails); err != nil {
			return errors.New("failed to create bom details")
		}
	} else {
		conditions := map[string]interface{}{
			"id": setupItemBomDetails.ID,
		}

		if err := services.DbUpdate(tx, &setupItemBomDetails, conditions); err != nil {
			return errors.New("failed updating child")
		}
	}

	setupItemBomDetailsAt := models.SetupItemBomDetailsAt{
		RefId:                      setupItemBomDetails.ID,
		SetupItemBomDetailsContent: setupItemBomDetails.SetupItemBomDetailsContent,
		At:                         at,
	}
	if err := services.DbInsert(tx, &setupItemBomDetailsAt); err != nil {
		return errors.New("failed creating SetupItemBomDetailsAt")
	}
	return nil
}

func GetBomDetails(child *[]models.SetupItemBomDetails, conditions map[string]interface{}) error {
	if err := services.DbGet(child, conditions); err != nil {
		return errors.New("failed getting bom details")
	}
	return nil
}
