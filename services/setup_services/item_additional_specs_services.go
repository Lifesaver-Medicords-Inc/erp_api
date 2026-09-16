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

	// Every additional-specs field the Item Entry form sends, blanks and zeros included (see
	// services.DbUpdateFields).
	fields := []string{"MaterialId", "SuctionPressure", "DriverType", "MotorEnclosure", "MotorManufacturer",
		"ServiceFactor", "LiquidType", "ConnectionType", "Size", "Volume", "VolumeUnitOfMeasureId",
		"Weight", "WeightUnitOfMeasureId", "Calibration", "LongDescription"}

	// Pump count and pump types are written only when the request actually carried them.
	// Item Entry used to send both under a misspelled key (…_compatability_id), so this model
	// never received them: pump count never saved, and the pump types were replaced with an
	// empty list - wiped - on every item update (user-reported 2026-09-14). The corrected app
	// sends the right keys; an app not yet reinstalled still sends the old ones, and for it
	// both are now left exactly as they are instead of being cleared.
	if additionalspec.PumpCountSent {
		fields = append(fields, "PumpCountCompatabilityId")
	}

	if err := services.DbUpdateFields(tx, &additionalspec.AdditionalSpecs, map[string]interface{}{"id": existing.ID}, fields...); err != nil {
		return errors.New("failed updating additional specs")
	}

	if additionalspec.PumpTypesSent {
		if err := UpdateAdditionalSpecsPumpType(tx, existing.ID, additionalspec.PumpTypeCompatabilityId, at); err != nil {
			return err
		}
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
