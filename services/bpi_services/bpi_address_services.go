package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateBpiAddress(tx *gorm.DB, parentId uint, general_id uint, child models.BpiAddress, salesId string, at models.At) error {

	child.BpiAddressContent.BasedId = parentId
	child.BpiAddressContent.BranchId = general_id

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to create bpi address")
	}

	childAt := models.BpiAddressAt{
		RefId:             child.ID,
		BpiAddressContent: child.BpiAddressContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi_address_at")
	}

	// create address history
	if err := CreateBpiHistory(tx, parentId, "create", "Address", salesId, at); err != nil {
		return err
	}

	return nil
}

func UpdateBpiAddress(tx *gorm.DB, generalId uint, child models.BpiAddress, salesId string, at models.At, conditions map[string]interface{}) error {

	if child.ID == 0 {
		// New record — insert instead of update
		child.BpiAddressContent.BasedId = conditions["based_id"].(uint)
		child.BpiAddressContent.BranchId = generalId
		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi address")
		}
	} else {
		// Existing record — fetch old for change detection, then update
		var oldAddress models.BpiAddress
		if err := tx.First(&oldAddress, child.ID).Error; err != nil {
			return errors.New("address record not found")
		}

		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi address")
		}

		var newAddress models.BpiAddress
		if err := tx.First(&newAddress, child.ID).Error; err != nil {
			return errors.New("failed re-fetching updated address")
		}

		if utils.HasChanged(oldAddress, newAddress) {
			if err := CreateBpiHistory(tx, generalId, "update", "Address", salesId, at); err != nil {
				return err
			}
		}
	}

	childAt := models.BpiAddressAt{
		RefId:             child.ID,
		BpiAddressContent: child.BpiAddressContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi address at in UpdateBpiAddress")
	}

	return nil
}
