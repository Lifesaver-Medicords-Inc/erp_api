package bpi_services

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type UpdateMainBranchDTO struct {
	ID     uint `json:"id"`
	IsMain bool `json:"is_main"`
}

func CreateBpiGeneral(tx *gorm.DB, parentId uint, child *models.BpiGeneralSchema, at models.At) error {

	child.BpiGeneralContent.BasedId = parentId

	// Backend owns these — strip whatever frontend sent
	child.BpiGeneralEmbeddedContent.CustomerCode = ""
	child.BpiGeneralEmbeddedContent.SupplierCode = ""
	

	if err := services.DbInsert(tx, &child.BpiGeneral); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("duplicate branch name")
		}
		return errors.New("failed creating branch")
	}

	childAt := models.BpiGeneralAt{
		RefId:                     child.ID,
		BranchName:                child.BranchName,
		SalesId:                   child.SalesId,
		IsMain:                    child.IsMain,
		BpiGeneralEmbeddedContent: child.BpiGeneralEmbeddedContent,
		At:                        at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi_generals_at")
	}

	if err := CreateBpiHistory(tx, child.BpiGeneral.ID, "create", "General", child.SalesId, at); err != nil {
		return err
	}

	for _, v := range child.BranchIndustryId {
		if err := CreateBpiBranchIndustries(tx, child.BpiGeneral.ID, uint(v), child.SalesId, at); err != nil {
			return err
		}
	}

	for _, v := range child.EntityTypeId {
		if err := CreateBpiEntity(tx, child.BpiGeneral.ID, uint(v), child.SalesId, at); err != nil {
			return err
		}
	}

	return nil
}

func UpdateBpiGeneral(tx *gorm.DB, child *models.BpiGeneralSchema, at models.At, conditions map[string]interface{}) error {

	var oldGeneral models.BpiGeneral
	if err := tx.First(&oldGeneral, child.BpiGeneral.ID).Error; err != nil {
		return err
	}

	// Preserve existing codes — backend owns these, ignore whatever frontend sent
	child.BpiGeneralEmbeddedContent.CustomerCode = oldGeneral.CustomerCode
	child.BpiGeneralEmbeddedContent.SupplierCode = oldGeneral.SupplierCode

	if err := services.DbUpdate(tx, &child.BpiGeneral, conditions); err != nil {
		return errors.New("failed to update bpi general")
	}

	generalat := models.BpiGeneralAt{
		RefId:                     child.ID,
		BranchName:                child.BranchName,
		SalesId:                   child.SalesId,
		IsMain:                    child.IsMain,
		BpiGeneralEmbeddedContent: child.BpiGeneralEmbeddedContent,
		At:                        at,
	}
	if err := services.DbInsert(tx, &generalat); err != nil {
		return errors.New("failed creating bpi_general_at")
	}

	var newGeneral models.BpiGeneral
	if err := tx.First(&newGeneral, child.BpiGeneral.ID).Error; err != nil {
		return err
	}

	if utils.HasChanged(oldGeneral, newGeneral) {
		if err := CreateBpiHistory(tx, child.BpiGeneral.ID, "update", "General", child.SalesId, at); err != nil {
			return err
		}
	}

	childConditions := map[string]interface{}{
		"bpi_general_id": child.BpiGeneral.ID,
	}

	if len(child.EntityTypeId) != 0 {
		if err := services.DbDelete(tx, &models.BpiEntity{}, childConditions); err != nil {
			return errors.New("failed to delete bpi entity")
		}
		for _, v := range child.EntityTypeId {
			if err := CreateBpiEntity(tx, child.BpiGeneral.ID, uint(v), child.SalesId, at); err != nil {
				return err
			}
		}
	}

	if len(child.BranchIndustryId) != 0 {
		if err := services.DbDelete(tx, &models.BpiBranchIndustries{}, childConditions); err != nil {
			return errors.New("failed to delete bpi industries")
		}
		for _, v := range child.BranchIndustryId {
			if err := CreateBpiBranchIndustries(tx, child.BpiGeneral.ID, uint(v), child.SalesId, at); err != nil {
				return err
			}
		}
	}

	return nil
}
