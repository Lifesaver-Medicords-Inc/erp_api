package bpi_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiEntity(tx *gorm.DB, parentId uint, entityId uint, at models.At) error {
	content := models.BpiEntityContent{
		BpiGeneralId: parentId,
		EntityId:     entityId,
	}
	bpiEntity := models.BpiEntity{BpiEntityContent: content}
	if err := services.DbInsert(tx, &bpiEntity); err != nil {
		return errors.New("failed creating bpi entity")
	}

	bpiEntityAt := models.BpiEntityAt{
		RefId:            bpiEntity.BpiGeneralId,
		BpiEntityContent: content,
		At:               at,
	}

	if err := services.DbInsert(tx, &bpiEntityAt); err != nil {
		return errors.New("failed creating bpi entity_at")
	}

	return nil
}
