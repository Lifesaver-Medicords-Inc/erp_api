package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiFinance(tx *gorm.DB, parentId uint, generalId uint, child models.BpiFinance, at models.At) error {

	child.BpiFinanceContent.FinanceBasedId = parentId
	child.BpiFinanceContent.FinanceBranchId = generalId
	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to create bpi finance")
	}
	childAt := models.BpiFinanceAt{
		RefId:             child.FinanceID,
		BpiFinanceContent: child.BpiFinanceContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi finance at")
	}

	return nil
}

func UpdateBpiFinance(tx *gorm.DB, child models.BpiFinance, at models.At, parentId uint) error {

	conditions := map[string]interface{}{
		"finance_based_id": parentId,
	}
	if err := services.DbUpdate(tx, &child, conditions); err != nil {
		return errors.New("failed updating bpi finance")
	}
	childAt := models.BpiFinanceAt{
		RefId:             child.FinanceID,
		BpiFinanceContent: child.BpiFinanceContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi finance at")
	}

	return nil
}
