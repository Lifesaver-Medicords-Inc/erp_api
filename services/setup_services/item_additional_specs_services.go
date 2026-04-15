package setup_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetAdditionalSpecs(additionalSpecs *[]models.AdditionalSpecsView, conditions map[string]interface{}) error {
	if err := services.DbGet(additionalSpecs, conditions); err != nil {
		fmt.Println("ERROR:", err)
		return errors.New("failed getting additional specs")
	}

	return nil
}

func GetAdditionalSpec(additionalspecs *models.AdditionalSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(additionalspecs, conditions); err != nil {
		return errors.New("failed getting additional spec")
	}

	return nil
}

func CreateAdditionalSpec(tx *gorm.DB, basedId uint, additionalSpec models.AdditionalSpecsSchema, at models.At) error {
	additionalSpec.BasedId = basedId

	if err := services.DbInsert(tx, &additionalSpec.AdditionalSpecs); err != nil {
		return errors.New("failed creating additional specs")
	}

	for _, v := range additionalSpec.PumpTypeCompatabilityId {
		if err := CreateAdditionalSpecsPumpType(tx, additionalSpec.ID, uint(v), at); err != nil {
			return err
		}
	}

	additionalSpecsAt := models.AdditionalSpecsAt{
		RefId:                  additionalSpec.ID,
		AdditionalSpecsContent: additionalSpec.AdditionalSpecsContent, // Snapshot of content
		At:                     at,
	}

	if err := services.DbInsert(tx, &additionalSpecsAt); err != nil {
		return errors.New("failed creating additional specs")
	}

	return nil
}

func UpdateAdditionalSpec(tx *gorm.DB, basedId uint, additionalspec models.AdditionalSpecsSchema, at models.At, conditions map[string]interface{}) error {
	var existing models.AdditionalSpecs

	err := services.DbGet(&existing, conditions)

	// ✅ If no existing record, create it using the real item ID
	if err != nil || existing.ID == 0 {
		return CreateAdditionalSpec(tx, basedId, additionalspec, at) // ✅ basedId from caller
	}

	// ✅ UPDATE: carry over the real ID and BasedId from DB
	additionalspec.ID = existing.ID
	additionalspec.BasedId = existing.BasedId

	if err := services.DbUpdate(tx, &additionalspec.AdditionalSpecs, map[string]interface{}{"id": existing.ID}); err != nil {
		return errors.New("failed updating additional specs")
	}

	if err := UpdateAdditionalSpecsPumpType(tx, existing.ID, additionalspec.PumpTypeCompatabilityId, at); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &models.AdditionalSpecsAt{
		RefId:                  existing.ID,
		AdditionalSpecsContent: additionalspec.AdditionalSpecsContent,
		At:                     at,
	}); err != nil {
		return errors.New("failed creating additional specs at")
	}

	return nil
}

func DeleteAdditionalSpecs(tx *gorm.DB, additionalspec models.AdditionalSpecs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &additionalspec, conditions); err != nil {
		return errors.New("failed deleting additional spec")
	}

	itemspecsat := models.AdditionalSpecsAt{
		RefId:                  additionalspec.BasedId,
		AdditionalSpecsContent: additionalspec.AdditionalSpecsContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &itemspecsat); err != nil {
		return errors.New("failed creating additional specsat")
	}

	return nil
}
