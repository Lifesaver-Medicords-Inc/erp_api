package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateBpiContact(tx *gorm.DB, parentId uint, generalId uint, child *models.BpiContacts, salesId string, at models.At) error {

	child.BpiContactContent.BasedId = parentId
	child.BpiContactContent.BranchId = generalId

	if err := services.DbInsert(tx, child); err != nil {
		return errors.New("failed to create bpi contacts")
	}

	childAt := models.BpiContactsAt{
		RefId:             child.ID,
		BpiContactContent: child.BpiContactContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi_contacts_at")
	}

	// create contact history
	if err := CreateBpiHistory(tx, parentId, "create", "Contacts", salesId, at); err != nil {
		return err
	}

	return nil
}

func UpdateBpiContact(tx *gorm.DB, generalId uint, child models.BpiContacts, salesId string, at models.At, conditions map[string]interface{}) error {

	if child.ID == 0 {
		// New record — insert
		child.BpiContactContent.BasedId = conditions["based_id"].(uint)
		child.BpiContactContent.BranchId = generalId
		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi contacts")
		}
	} else {
		// Existing record — fetch old, update, compare
		var oldContacts models.BpiContacts
		if err := tx.First(&oldContacts, child.ID).Error; err != nil {
			return errors.New("contact record not found")
		}

		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi contacts")
		}

		var newContacts models.BpiContacts
		if err := tx.First(&newContacts, child.ID).Error; err != nil {
			return errors.New("failed re-fetching updated contact")
		}

		if utils.HasChanged(oldContacts, newContacts) {
			if err := CreateBpiHistory(tx, generalId, "update", "Contacts", salesId, at); err != nil {
				return err
			}
		}
	}

	childAt := models.BpiContactsAt{
		RefId:             child.ID,
		BpiContactContent: child.BpiContactContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi contact at in UpdateBpiContact")
	}

	return nil
}
