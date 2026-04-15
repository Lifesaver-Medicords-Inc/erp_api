package bpi_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateBpiFinance(tx *gorm.DB, parentId uint, generalId uint, child models.BpiFinance, salesId string, at models.At) error {
	child.FinanceBasedId = parentId
	child.FinanceBranchId = generalId

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

	// create finance history
	if err := CreateBpiHistory(tx, parentId, "create", "Finance", salesId, at); err != nil {
		return err
	}

	return nil
}

func UpdateBpiFinance(tx *gorm.DB, child models.BpiFinance, salesId string, at models.At, parentId uint) error {
	fmt.Println("BPI FINANCE: ", child)

	if child.FinanceID == 0 {
		// New record — insert
		child.FinanceBasedId = parentId
		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi finance")
		}
	} else {
		// Existing record — fetch old, update, compare
		oldFinance := models.BpiFinance{}
		if err := tx.First(&oldFinance, child.FinanceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		conditions := map[string]interface{}{
			"finance_id": child.FinanceID,
		}
		fmt.Println("BPI FINANCE CONDITIONS: ", conditions)
		if err := services.DbUpdate(tx, &child, conditions); err != nil {
			return errors.New("failed updating bpi finance")
		}

		newFinance := models.BpiFinance{}
		if err := tx.First(&newFinance, child.FinanceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if utils.HasChanged(oldFinance, newFinance) {
			if err := CreateBpiHistory(tx, parentId, "update", "Finance", salesId, at); err != nil {
				return err
			}
		}
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
