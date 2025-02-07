package bpi_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiContact(tx *gorm.DB, parentId uint, child models.BpiContacts, at models.At) error {

	child.BpiContactContent.BasedId = parentId

	fmt.Println("BPI CONTACT1234")
	if err := services.DbInsert(tx, &child); err != nil {
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
