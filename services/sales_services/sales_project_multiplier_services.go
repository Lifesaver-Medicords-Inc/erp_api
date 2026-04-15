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

func UpdateSalesProjectMultiplier(tx *gorm.DB, projectmultiplier models.SalesProjectMultiplier, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectmultiplier, conditions); err != nil {
		return errors.New("failed updating project multiplier")
	}

	projectmultiplierat := models.SalesProjectMultiplierAt{
		RefId:                         projectmultiplier.MultiplierID,
		SalesProjectMultiplierContent: projectmultiplier.SalesProjectMultiplierContent,
		At:                            at,
	}

	if err := services.DbInsert(tx, &projectmultiplierat); err != nil {
		return errors.New("failed creating project multiplier at")
	}

	return nil
}
