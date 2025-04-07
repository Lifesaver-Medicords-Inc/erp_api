package bpi_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateBpiEntity(tx *gorm.DB, parentId uint, entityId uint, at models.At) error {

	fmt.Println("CREATE BPI ENTITTY")
	content := models.BpiEntityContent{
		BpiGeneralId: parentId,
		EntityId:     entityId,
	}
	bpiEntity := models.BpiEntity{BpiEntityContent: content}
	fmt.Println("CREATE BPI ENTITTYs", bpiEntity.BpiEntityContent)

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

func UpdateBpiEntity(tx *gorm.DB, parentId uint, entityId uint, at models.At) error {

	conditions := map[string]interface{}{
		"bpi_general_id": parentId,
		"entity_id":      entityId,
	}

	fmt.Println("ENTITY ID", entityId)

	content := models.BpiEntityContent{
		EntityId: entityId,
	}

	bpiEntity := models.BpiEntity{BpiEntityContent: content}
	if err := services.DbUpdate(tx, &bpiEntity, conditions); err != nil {
		return errors.New("failed updating bpi entity")
	}

	bpiEntityAt := models.BpiEntityAt{
		RefId:            bpiEntity.BpiGeneralId,
		BpiEntityContent: content,
		At:               at,
	}

	if err := services.DbInsert(tx, &bpiEntityAt); err != nil {
		return errors.New("failed creating bpi_industries_at")
	}

	return nil
}
