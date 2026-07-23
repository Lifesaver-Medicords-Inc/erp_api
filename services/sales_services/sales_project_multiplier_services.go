package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// CREATE MULTIPLIERS
func CreateSalesProjectMultiplier(tx *gorm.DB, parentId uint, multiplier models.SalesProjectMultiplier, at models.At) error {
	multiplier.BasedId = parentId

	// Same fix as CreateProjectItemSet's ItemSetID = 0 - MultiplierID is client-controlled on
	// the wire, but this function only ever creates a brand new row. Always let the DB assign
	// a fresh id instead of trusting whatever (possibly stale/leftover) id the client sent.
	multiplier.MultiplierID = 0

	if err := services.DbInsert(tx, &multiplier); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &multiplier)
		return errors.New("failed creating multiplier")
	}

	multiplierat := models.SalesProjectMultiplierAt{
		RefId:                         multiplier.MultiplierID,
		SalesProjectMultiplierContent: multiplier.SalesProjectMultiplierContent,
		At:                            at,
	}

	if err := services.DbInsert(tx, &multiplierat); err != nil {
		return errors.New("failed creating multipliers")
	}

	return nil
}

// GET MULTIPLIERS
func GetSalesProjectMultiplier(multiplier *[]models.SalesProjectMultiplier, conditions map[string]interface{}) error {
	if err := services.DbGet(multiplier, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}

func applyMultiplierDiff(tx *gorm.DB, Id uint, diff CollectionDiff[models.SalesProjectMultiplier], at models.At) error {

	// ---- ADDED ----
	for _, item := range diff.Added {
		item.BasedId = Id
		// Same reason as CreateProjectContent's ContentID = 0 - always let the DB assign a
		// fresh id for anything landing in Added, never trust a client-sent MultiplierID.
		item.MultiplierID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("multiplier add: %w", err)
		}

		atRecord := models.SalesProjectMultiplierAt{
			RefId:                         item.MultiplierID,
			SalesProjectMultiplierContent: item.SalesProjectMultiplierContent,
			At:                            at,
		}
		if err := services.DbInsert(tx, &atRecord); err != nil {
			return fmt.Errorf("multiplier add at: %w", err)
		}
	}

	// ---- REMOVED ----
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectMultiplier{}, map[string]interface{}{
			"multiplier_id": item.MultiplierID,
		}); err != nil {
			return fmt.Errorf("multiplier remove: %w", err)
		}
	}

	// ---- UPDATED ----
	for _, entry := range diff.Updated {
		if err := UpdateSalesProjectMultiplier(tx, entry.Item, at, map[string]interface{}{
			"multiplier_id": entry.Item.MultiplierID,
		}); err != nil {
			return fmt.Errorf("multiplier update: %w", err)
		}
	}

	return nil
}

func UpdateSalesProjectMultiplier(tx *gorm.DB, projectmultiplier models.SalesProjectMultiplier, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectmultiplier, conditions); err != nil {
		return errors.New("failed updating project multiplier")
	}

	projectmultiplierat := models.SalesProjectMultiplierAt{
		RefId:                         projectmultiplier.MultiplierID,
		SalesProjectMultiplierContent: projectmultiplier.SalesProjectMultiplierContent,
		At:                            at,
	}

	// DbInsert here will duplicate the At record every update —
	// should be DbUpdate keyed on ref_id instead
	if err := services.DbUpdate(tx, &projectmultiplierat, map[string]interface{}{
		"ref_id": projectmultiplier.MultiplierID,
	}); err != nil {
		return errors.New("failed updating project multiplier at")
	}

	return nil
}
