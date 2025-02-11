package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiItems(tx *gorm.DB, parentId uint, childItems []models.BpiItems, at models.At) error {

	for _, v := range childItems {
		if err := CreateBpiItem(tx, parentId, v, at); err != nil {
			return err
		}
	}

	return nil

}

func CreateBpiItem(tx *gorm.DB, parentId uint, child models.BpiItems, at models.At) error {

	child.BpiItemContent.BasedId = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to create bpi items")
	}
	childAt := models.BpiItemsAt{
		RefId:          child.ID,
		BpiItemContent: child.BpiItemContent,
		At:             at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi items at")
	}

	return nil
}
