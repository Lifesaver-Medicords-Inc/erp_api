package bpi_services

import (
	"errors"
	"time"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiHistory(tx *gorm.DB, basedId uint, action string, childType string, editBy string, at models.At) error {
	history := models.BpiHistory{
		BpiHistoryContent: models.BpiHistoryContent{
			BasedId:   basedId,
			Date:      time.Now().Format("2006-01-02 15:04:05"),
			Actions:   action,
			ChildType: childType,
			EditBy:    editBy,
		},
	}

	if err := services.DbInsert(tx, &history); err != nil {
		return errors.New("failed creating bpi history")
	}

	historyAt := models.BpiHistoryAt{
		RefId:             history.ID,
		BpiHistoryContent: history.BpiHistoryContent,
		At:                at,
	}

	if err := services.DbInsert(tx, &historyAt); err != nil {
		return errors.New("failed creating bpi_history_at")
	}

	return nil
}
