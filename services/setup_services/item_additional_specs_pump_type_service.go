package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateAdditionalSpecsPumpType(tx *gorm.DB, basedId uint, pumpTypeCompatabilityId uint, at models.At) error {
	content := models.AdditionalSpecsPumpTypeContent{
		AdditionalSpecsId:       basedId,
		PumpTypeCompatabilityId: pumpTypeCompatabilityId,
	}
	additionalSpecsPumpType := models.AdditionalSpecsPumpType{AdditionalSpecsPumpTypeContent: content}
	if err := services.DbInsert(tx, &additionalSpecsPumpType); err != nil {
		return errors.New("failed creating additional specs pump type")
	}

	additionalSpecsPumpTypeAt := models.AdditionalSpecsPumpTypeAt{
		RefId:                          additionalSpecsPumpType.ID,
		AdditionalSpecsPumpTypeContent: content,
		At:                             at,
	}

	if err := services.DbInsert(tx, &additionalSpecsPumpTypeAt); err != nil {
		return errors.New("failed creating additional specs pump type at")
	}

	return nil
}

func UpdateAdditionalSpecsPumpType(tx *gorm.DB, additionalSpecsID uint, pumpTypeIDs []uint, at models.At) error {
	if err := services.DbDelete(tx, &models.AdditionalSpecsPumpType{}, map[string]interface{}{"additional_specs_id": additionalSpecsID}); err != nil {
		return errors.New("failed deleting existing pump additional specs pump type")
	}

	for _, pumpTypeID := range pumpTypeIDs {
		if err := CreateAdditionalSpecsPumpType(tx, additionalSpecsID, pumpTypeID, at); err != nil {
			return err
		}
	}

	return nil
}
