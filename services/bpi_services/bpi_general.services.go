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
	child.BasedId = parentId

	// Backend owns these — strip whatever frontend sent
	child.CustomerCode = ""
	child.SupplierCode = ""

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

	if err := CreateBpiHistory(tx, child.ID, "create", "General", child.SalesId, at); err != nil {
		return err
	}

	for _, v := range child.BranchIndustryId {
		if err := CreateBpiBranchIndustries(tx, child.ID, uint(v), child.SalesId, at); err != nil {
			return err
		}
	}

	for _, v := range child.EntityTypeId {
		if err := CreateBpiEntity(tx, child.ID, uint(v), child.SalesId, at); err != nil {
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
	child.CustomerCode = oldGeneral.CustomerCode
	child.SupplierCode = oldGeneral.SupplierCode

	// An edit never changes who owns the branch (spec 4.1.10, 14.168): ownership
	// moves only through an explicit reassignment. The client sends the EDITOR's
	// name in sales_id, and every history row - the ones below, and the contact /
	// address / item / finance rows UpdateBpi writes after this returns - records
	// it as edit_by. So child.SalesId stays the editor, and only the row actually
	// saved carries the stored owner. Done here rather than in the client so it
	// holds for every client build, including ones already installed.
	saved := child.BpiGeneral
	saved.SalesId = oldGeneral.SalesId

	if err := services.DbUpdate(tx, &saved, conditions); err != nil {
		return errors.New("failed to update bpi general")
	}

	generalat := models.BpiGeneralAt{
		RefId:                     child.ID,
		BranchName:                child.BranchName,
		SalesId:                   saved.SalesId,
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
		if err := CreateBpiHistory(tx, child.ID, "update", "General", child.SalesId, at); err != nil {
			return err
		}
	}

	childConditions := map[string]interface{}{
		"bpi_general_id": child.ID,
	}

	if len(child.EntityTypeId) != 0 {
		if err := services.DbDelete(tx, &models.BpiEntity{}, childConditions); err != nil {
			return errors.New("failed to delete bpi entity")
		}
		for _, v := range child.EntityTypeId {
			if err := CreateBpiEntity(tx, child.ID, uint(v), child.SalesId, at); err != nil {
				return err
			}
		}
	}

	if len(child.BranchIndustryId) != 0 {
		if err := services.DbDelete(tx, &models.BpiBranchIndustries{}, childConditions); err != nil {
			return errors.New("failed to delete bpi industries")
		}
		for _, v := range child.BranchIndustryId {
			if err := CreateBpiBranchIndustries(tx, child.ID, uint(v), child.SalesId, at); err != nil {
				return err
			}
		}
	}

	return nil
}
