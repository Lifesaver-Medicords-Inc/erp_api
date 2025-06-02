package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiAddress(tx *gorm.DB, parentId uint, general_id uint, child models.BpiAddress, at models.At) error {

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

	return nil
}

func UpdateBpiAddress(tx *gorm.DB, generalId uint, child models.BpiAddress, at models.At, conditions map[string]interface{}) error {

	if child.ID == 0 {
		child.BpiAddressContent.BasedId = conditions["based_id"].(uint)
		child.BpiAddressContent.BranchId = generalId
		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi address")
		}

	} else {

		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi address")
		}
	}

	childfat := models.BpiAddressAt{
		RefId:             child.ID,
		BpiAddressContent: child.BpiAddressContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating bpi address at  in Update Bpi adresss")
	}
	return nil
}
