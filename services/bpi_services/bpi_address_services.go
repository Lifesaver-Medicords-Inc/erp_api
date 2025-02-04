package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiAddress(tx *gorm.DB, parentId uint, child models.BpiAddress, at models.At) error {

	child.BpiAddressContent.BasedId = parentId
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
