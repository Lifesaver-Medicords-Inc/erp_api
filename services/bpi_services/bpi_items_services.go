package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateBpiItems(tx *gorm.DB, parentId uint, generalId uint, childItems []models.BpiItems, salesId string, at models.At) error {

	for _, v := range childItems {
		if err := CreateBpiItem(tx, parentId, generalId, v, salesId, at); err != nil {
			return err
		}
	}
	return nil
}

func CreateBpiItem(tx *gorm.DB, parentId uint, generalId uint, child models.BpiItems, salesId string, at models.At) error {

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

	// create items history
	if err := CreateBpiHistory(tx, parentId, "create", "Items", salesId, at); err != nil {
		return err
	}

	return nil
}

func UpdateBpiItems(tx *gorm.DB, parentId uint, generalId uint, childItems []models.BpiItems, salesId string, at models.At) error {

	var existingIDs []uint
	if err := tx.
		Model(&models.BpiItems{}).
		Where("based_id = ?", parentId).
		Pluck("id", &existingIDs).Error; err != nil {
		return err
	}

	// 2. Build incoming ID map
	incomingMap := make(map[uint]bool)
	for _, v := range childItems {
		if v.ID > 0 {
			incomingMap[v.ID] = true
		}
	}

	// 3. DELETE items removed from the list
	for _, id := range existingIDs {
		if !incomingMap[id] {
			if err := tx.Delete(&models.BpiItems{}, id).Error; err != nil {
				return err
			}

			// history
			if err := CreateBpiHistory(tx, parentId, "delete", "Items", salesId, at); err != nil {
				return err
			}
		}
	}

	// 4. CREATE / UPDATE
	for _, v := range childItems {
		if v.ID == 0 {
			// CREATE
			if err := CreateBpiItem(tx, parentId, generalId, v, salesId, at); err != nil {
				return err
			}
		} else {
			// UPDATE
			if err := UpdateBpiItem(tx, parentId, generalId, v, salesId, at); err != nil {
				return err
			}
		}
	}

	return nil
}

func UpdateBpiItem(tx *gorm.DB, parentId uint, generalId uint, child models.BpiItems, salesId string, at models.At) error {

	oldItem := models.BpiItems{}
	if err := tx.First(&oldItem, child.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

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

	newItem := models.BpiItems{}
	if err := tx.First(&newItem, child.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if utils.HasChanged(oldItem, newItem) {
		// create items history
		if err := CreateBpiHistory(tx, parentId, "update", "Items", salesId, at); err != nil {
			return err
		}
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
func DeleteBpiItem( tx *gorm.DB, parentId uint, itemId uint, salesId string, at models.At,) error {

	
	if err := tx.
		Where("ref_id = ?", itemId).
		Delete(&models.BpiItemsAt{}).Error; err != nil {
		return errors.New("failed deleting bpi item at")
	}

	// delete main item
	if err := tx.
		Delete(&models.BpiItems{}, itemId).Error; err != nil {
		return errors.New("failed deleting bpi item")
	}

	// history
	if err := CreateBpiHistory(tx, parentId, "delete", "Items", salesId, at); err != nil {
		return err
	}

	return nil
}
