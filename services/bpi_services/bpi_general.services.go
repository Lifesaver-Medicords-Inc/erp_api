package bpi_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func BpiGeneral(tx *gorm.DB, parentId uint, child *models.BpiGeneralSchema, at models.At) error {

	fmt.Println("BPI GENERAL ")

	child.BpiGeneralContent.BasedId = parentId
	if err := services.DbInsert(tx, &child.BpiGeneral); err != nil {

		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate branch name")
		} else {
			err = errors.New("failed creating brand")
		}

		return err
	}

	for _, v := range child.BranchIndustryId {
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

func UpdateBpiGeneral(tx *gorm.DB, child models.BpiGeneralSchema, at models.At, conditions map[string]interface{}) (uint, error) {

	childConditions := map[string]interface{}{
		"bpi_general_id": child.BpiGeneral.ID,
	}

	if err := services.DbUpdate(tx, &child.BpiGeneral, conditions); err != nil {
		return 0, errors.New("failed to update bpi general")
	}

	//Check if entity id has data to delete then Create new Data in entity table
	if len(child.EntityTypeId) != 0 {
		fmt.Println("ENTITY TYPE", child.EntityTypeId)

		if err := services.DbDelete(tx, &models.BpiEntity{}, childConditions); err != nil {
			return 0, errors.New("failed to delete bpi entity")
		}
	}

	//Check if branch industry id has data to delete then Create new Data in branch industries  table
	if len(child.BranchIndustryId) != 0 {
		fmt.Println("BRANCH INDUSTRY TYPE", child.BranchIndustryId)
		if err := services.DbDelete(tx, &models.BpiBranchIndustries{}, childConditions); err != nil {
			return 0, errors.New("failed to delete bpi industry")
		}
	}

	for _, v := range child.BranchIndustryId {
		fmt.Println("Branch Industriess", v)
		if err := CreateBpiBranchIndustries(tx, child.BpiGeneral.ID, uint(v), at); err != nil {
			return 0, err
		}
	}

	for _, v := range child.EntityTypeId {
		fmt.Println("Entity", v)
		if err := CreateBpiEntity(tx, child.BpiGeneral.ID, uint(v), at); err != nil {
			return 0, err
		}
	}

	return child.ID, nil
}
