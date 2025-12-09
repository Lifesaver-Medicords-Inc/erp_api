package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiHistoryDetails(tx *gorm.DB, basedId uint, branchId uint, columnName string, oldValue string, newValue string, at models.At) error {

	historyDetails := models.BpiHistoryDetails{
		BpiHistoryDetailsContent: models.BpiHistoryDetailsContent{
			BasedId:  basedId,
			BranchId: branchId,
			Name:     columnName,
			OldValue: oldValue,
			NewValue: newValue,
		},
	}

	if err := services.DbInsert(tx, &historyDetails); err != nil {
		return errors.New("failed creating bpi history details")
	}

	historyDetailsAt := models.BpiHistoryDetailsAt{
		RefId:                     historyDetails.ID,
		BpiHistoryDetailsContent: historyDetails.BpiHistoryDetailsContent,
		At:                        at,
	}	

	if err := services.DbInsert(tx, &historyDetailsAt); err != nil {
		return errors.New("failed creating bpi_history_details_at")
	}
	return nil
}
