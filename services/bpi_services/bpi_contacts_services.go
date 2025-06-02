package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiContact(tx *gorm.DB, parentId uint, generalId uint, child *models.BpiContacts, at models.At) error {

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

	return nil
}

func UpdateBpiContact(tx *gorm.DB, generalId uint, child models.BpiContacts, at models.At, conditions map[string]interface{}) error {

	if child.ID == 0 {
		child.BpiContactContent.BasedId = conditions["based_id"].(uint)
		child.BpiContactContent.BranchId = generalId

		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi contacts")
		}

	} else {
		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi contacts")
		}
	}

	childfat := models.BpiContactsAt{
		RefId:             child.ID,
		BpiContactContent: child.BpiContactContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating bpi contact at  in Update Bpi Contacts")
	}
	return nil
}
