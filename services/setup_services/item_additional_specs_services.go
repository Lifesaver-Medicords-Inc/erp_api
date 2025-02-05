package setup_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetAdditionalSpecs(additionalSpecs *[]models.AdditionalSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(additionalSpecs, conditions); err != nil {
		fmt.Println("ERROR:", err)
		return errors.New("failed getting item specs")
	}

	return nil
}

func GetAdditionalSpec(additionalspecs *models.AdditionalSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(additionalspecs, conditions); err != nil {
		return errors.New("failed getting additional spec")
	}

	return nil
}

func CreateAdditionalSpec(tx *gorm.DB, basedId uint, additionalSpec models.AdditionalSpecs, at models.At) error {
	content := models.AdditionalSpecsContent{
		BasedId:           basedId,
		SuctionPressure:   additionalSpec.SuctionPressure,
		DriverType:        additionalSpec.DriverType,
		MotorEnclosure:    additionalSpec.MotorEnclosure,
		MotorManufacturer: additionalSpec.MotorManufacturer,
		ServiceFactor:     additionalSpec.ServiceFactor,
		LiquidType:        additionalSpec.LiquidType,
		Volume:            additionalSpec.Volume,
		Weight:            additionalSpec.Weight,
		LongDescription:   additionalSpec.LongDescription,
	}

	additionalSpecs := models.AdditionalSpecs{AdditionalSpecsContent: content}
	if err := services.DbInsert(tx, &additionalSpecs); err != nil {
		return fmt.Errorf("failed creating additional specs")
	}

	additionalSpecsAt := models.AdditionalSpecsAt{
		RefId:                  additionalSpecs.ID,
		AdditionalSpecsContent: content,
		At:                     at,
	}

	if err := services.DbInsert(tx, &additionalSpecsAt); err != nil {
		return errors.New("failed creating additional specs")
	}

	return nil
}

func UpdateAdditionalSpec(tx *gorm.DB, additionalspec models.AdditionalSpecs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &additionalspec, conditions); err != nil {
		return errors.New("failed updating additional specs")
	}

	additionalspecat := models.AdditionalSpecsAt{
		RefId:                  additionalspec.ID,
		AdditionalSpecsContent: additionalspec.AdditionalSpecsContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &additionalspecat); err != nil {
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
