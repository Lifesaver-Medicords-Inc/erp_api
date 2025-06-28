package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiItems(tx *gorm.DB, parentId uint, generalId uint, childItems []models.BpiItems, at models.At) error {

	for _, v := range childItems {
		if err := CreateBpiItem(tx, parentId, generalId, v, at); err != nil {
			return err
		}
	}
	return nil
}

func CreateBpiItem(tx *gorm.DB, parentId uint, generalId uint, child models.BpiItems, at models.At) error {

	child.BpiItemContent.BasedId = parentId
	child.BpiItemContent.BranchId = generalId

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

func UpdateBpiItems(tx *gorm.DB, parentId uint, generalId uint, childItems []models.BpiItems, at models.At) error {

	for _, v := range childItems {
		if err := UpdateBpiItem(tx, parentId, generalId, v, at); err != nil {
			return err
		}
	}

	return nil
}

func UpdateBpiItem(tx *gorm.DB, parentId uint, generalId uint, child models.BpiItems, at models.At) error {

	conditions := map[string]interface{}{
		"id": child.ID,
	}
	
	if child.ID == 0 {
		child.BpiItemContent.BasedId = parentId
		child.BpiItemContent.BranchId = generalId

		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi items")
		}

	} else {
		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi items")
		}
	}

	childfat := models.BpiItemsAt{
		RefId:          child.ID,
		BpiItemContent: child.BpiItemContent,
		At:             at,
	}

	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating bpi items at  in Update Bpi items")
	}
	return nil
}

func UpdateBpiItemCanvass(tx *gorm.DB, canvassId uint, price float64, conditions map[string]interface{}, at models.At) error {

	var body models.BpiItems

	// First clause to call for specific record, for the mean time use find to get all that checks the condition
	if err := tx.Model(&models.BpiItems{}).Where(conditions).Find(&body).Error; err != nil {
		return errors.New("failed getting existing Bpi items")
	}
	body.Price = price
	body.CanvassId = canvassId

	// where's the part where price will be updated
	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return errors.New("failed to update bpi item canvass")
	}

	childat := models.BpiItemsAt{
		RefId:          body.ID,
		BpiItemContent: body.BpiItemContent,
		At:             at,
	}

	if err := services.DbInsert(tx, &childat); err != nil {
		return errors.New("failed to creating bpi items at in Update Bpi items canvass")
	}

	return nil
}
