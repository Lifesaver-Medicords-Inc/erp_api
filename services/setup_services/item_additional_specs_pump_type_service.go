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

func UpdateAdditionalSpecsPumpType(tx *gorm.DB, body SaveBody, at models.At) error {
	if err := services.DbDelete(tx, &models.AdditionalSpecsPumpType{}, map[string]interface{}{"based_id": body.ID}); err != nil {
		return errors.New("failed deleting existing additional specs pump type")
	}

	for _, v := range body.PumpTypeId {
		if err := CreateAdditionalSpecsPumpType(tx, body.ID, v, at); err != nil {
			return err
		}
	}

	return nil
}
