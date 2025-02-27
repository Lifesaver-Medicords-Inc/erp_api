package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BomDetails(tx *gorm.DB, parentId uint, child models.SetupItemBomDetails, at models.At) error {

	child.SetupItemBomDetailsContent.ItemBomID = parentId

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
