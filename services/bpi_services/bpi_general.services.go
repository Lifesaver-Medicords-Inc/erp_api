package bpi_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiGeneral(tx *gorm.DB, parentId uint, child models.BpiGeneralSchema, at models.At) error {

	fmt.Println("BPI GENERAL ")
	child.BpiGeneralContent.BasedId = parentId
	if err := services.DbInsert(tx, &child.BpiGeneral); err != nil {
		return errors.New("failed to create bpi general")
	}

	for _, v := range child.BranchIndustriesId {
		fmt.Println("Branch Industriess", v)
		if err := CreateBpiBranchIndustries(tx, child.BpiGeneral.ID, uint(v), at); err != nil {
			return err
		}
	}
	for _, v := range child.EntityTypeId {
		fmt.Println("Entity", v)
		if err := CreateBpiEntity(tx, child.BpiGeneral.ID, uint(v), at); err != nil {
			return err
		}
	}

	return nil
}
