package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiIndustries(tx *gorm.DB, parentId uint, industryId uint, at models.At) error {

	content := models.BpiIndustriesContent{
		BpiId:      parentId,
		IndustryId: industryId,
	}
	bpiIndustries := models.BpiIndustries{BpiIndustriesContent: content}
	if err := services.DbInsert(tx, &bpiIndustries); err != nil {
		return errors.New("failed creating bpi industries")
	}

	bpiIndustriesAt := models.BpiIndustriesAt{
		RefId:                bpiIndustries.ID,
		BpiIndustriesContent: content,
		At:                   at,
	}

	if err := services.DbInsert(tx, &bpiIndustriesAt); err != nil {
		return errors.New("failed creating bpi_industries_at")
	}

	return nil
}

func UpdateBpiIndustries(tx *gorm.DB, parentId uint, industryId uint, at models.At) error {

	content := models.BpiIndustriesContent{
		BpiId:      parentId,
		IndustryId: industryId,
	}
	bpiIndustries := models.BpiIndustries{BpiIndustriesContent: content}
	if err := services.DbInsert(tx, &bpiIndustries); err != nil {
		return errors.New("failed creating bpi industries")
	}

	if err := services.DbInsert(tx, &bpiIndustries); err != nil {
		return errors.New("failed creating bpi industries")
	}

	bpiIndustriesAt := models.BpiIndustriesAt{
		RefId:                bpiIndustries.ID,
		BpiIndustriesContent: content,
		At:                   at,
	}

	if err := services.DbInsert(tx, &bpiIndustriesAt); err != nil {
		return errors.New("failed creating bpi_industries_at")
	}

	return nil
}

func CreateBpiBranchIndustries(tx *gorm.DB, parentId uint, branchIndustryId uint, salesId string, at models.At) error {

	content := models.BpiBranchIndustriesContent{
		BpiGeneralId: parentId,
		IndustryId:   branchIndustryId,
	}
	bpiBranchIndustries := models.BpiBranchIndustries{BpiBranchIndustriesContent: content}
	if err := services.DbInsert(tx, &bpiBranchIndustries); err != nil {
		return errors.New("failed creating bpi branch industries")
	}

	bpiBranchIndustriesAt := models.BpiBranchIndustriesAt{
		RefId:                      bpiBranchIndustries.BpiGeneralId,
		BpiBranchIndustriesContent: content,
		At:                         at,
	}

	if err := services.DbInsert(tx, &bpiBranchIndustriesAt); err != nil {
		return errors.New("failed creating bpi branch_industries_at")
	}

	// create branch industries history
	// if err := CreateBpiHistory(tx, parentId, "create", "Branch Industries", salesId, at); err != nil {
	// 	return err
	// }

	return nil
}

func UpdateBpiBranchIndustries(tx *gorm.DB, parentId uint, branchIndustryId uint, salesId string, at models.At) error {

	conditions := map[string]interface{}{
		"bpi_general_id": parentId,
		"industry_id":    branchIndustryId,
	}

	content := models.BpiBranchIndustriesContent{
		IndustryId: branchIndustryId,
	}
	bpiBranchIndustries := models.BpiBranchIndustries{BpiBranchIndustriesContent: content}
	if err := services.DbUpdate(tx, &bpiBranchIndustries, conditions); err != nil {
		return errors.New("failed updating bpi branch  industries")
	}

	bpiBranchIndustriesAt := models.BpiBranchIndustriesAt{
		RefId:                      bpiBranchIndustries.BpiGeneralId,
		BpiBranchIndustriesContent: content,
		At:                         at,
	}

	if err := services.DbInsert(tx, &bpiBranchIndustriesAt); err != nil {
		return errors.New("failed creating bpi branch industries at")
	}

	// create branch industries history
	if err := CreateBpiHistory(tx, parentId, "update", "Branch Industries", salesId, at); err != nil {
		return err
	}

	return nil
}
