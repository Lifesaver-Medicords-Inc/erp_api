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

func CreateBpiBranchIndustries(tx *gorm.DB, parentId uint, branchIndustryId uint, at models.At) error {

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

	return nil
}
