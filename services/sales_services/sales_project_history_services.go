package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateSalesProjectHistory(tx *gorm.DB, parentId uint, history models.SalesProjectHistory, at models.At) error {

	history.BasedId = parentId

	if err := services.DbInsert(tx, &history); err != nil {
		return errors.New("failed creating history")
	}

	historyat := models.SalesProjectHistoryAt{
		RefId:                      history.HistoryID,
		SalesProjectHistoryContent: history.SalesProjectHistoryContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &historyat); err != nil {
		return errors.New("failed creating history")
	}

	return nil
}

func GetSalesProjectHistory(history *[]models.SalesProjectHistory, conditions map[string]interface{}) error {
	if err := services.DbGet(history, conditions); err != nil {
		return errors.New("failed getting history")
	}
	return nil
}
