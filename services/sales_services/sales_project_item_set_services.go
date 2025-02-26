package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItemSet(tx *gorm.DB, parentID uint, ItemSet *models.SalesProjectItemSet, at models.At) error {

	ItemSet.BasedId = parentID

	fmt.Println("LLLLLLLLLLL")

	if err := services.DbInsert(tx, ItemSet); err != nil {
		return errors.New("failed creating item set")
	}

	fmt.Println("KKKKKKKKKKKK")

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      ItemSet.ID,
		SalesProjectItemSetContent: ItemSet.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil
}

func CreateSeparateProjectItemSet(tx *gorm.DB, ItemSet *models.SalesProjectItemSet, at models.At) error {

	if err := services.DbInsert(tx, &ItemSet); err != nil {
		return errors.New("failed creating item set")
	}

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      ItemSet.ID,
		SalesProjectItemSetContent: ItemSet.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil
}

func GetProjectItemSet(itemset *[]models.SalesProjectItemSet, conditions map[string]interface{}) error {
	if err := services.DbGet(itemset, conditions); err != nil {
		return errors.New("failed getting itemsets")
	}

	return nil
}

// DO UDPATE SOON

func UpdateProjectItemSet(tx *gorm.DB, itemset models.SalesProjectItemSet, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &itemset, conditions); err != nil {
		return errors.New("failed updating project item set")
	}

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      itemset.ID,
		SalesProjectItemSetContent: itemset.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil

}
