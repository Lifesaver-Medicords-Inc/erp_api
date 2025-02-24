package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItemSet(tx *gorm.DB, parentID uint, ItemSet *models.SalesProjectItemSet, at models.At) error {

	ItemSet.BasedId = parentID
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
