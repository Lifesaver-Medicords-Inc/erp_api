package setup_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetAdditionalSpecs(additionalspecs *[]models.AdditionalSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(additionalspecs, conditions); err != nil {
		fmt.Println("ERRRRR:", err)
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

func Createadditionalspec(tx *gorm.DB, basedId uint, additionalspec models.AdditionalSpecs, at models.At) error {
	content := models.AdditionalSpecsContent{
		BasedId:           basedId,
		ItemCode:          additionalspec.ItemCode,
		SuctionPressure:   additionalspec.SuctionPressure,
		DriverType:        additionalspec.DriverType,
		MotorEnclosure:    additionalspec.MotorEnclosure,
		MotorManufacturer: additionalspec.MotorManufacturer,
		ServiceFactor:     additionalspec.ServiceFactor,
		Liquidtype:        additionalspec.Liquidtype,
		Volume:            additionalspec.Volume,
		Weight:            additionalspec.Weight,
		LongDescription:   additionalspec.LongDescription,
	}

	additionalSpecs := models.AdditionalSpecs{AdditionalSpecsContent: content}
	if err := services.DbInsert(tx, &additionalSpecs); err != nil {
		return fmt.Errorf("failed creating additional specs")
	}
	additionalspecsat := models.AdditionalSpecsAt{
		RefId:                  additionalSpecs.BasedId,
		AdditionalSpecsContent: content,
		At:                     at,
	}
	if err := services.DbInsert(tx, &additionalspecsat); err != nil {
		return errors.New("failed creating additional specs")
	}

	return nil
}

func UpdateAdditionalSpecs(tx *gorm.DB, additionalspec models.AdditionalSpecs, at models.At) error {
	if err := services.DbUpdate(tx, &additionalspec, nil); err != nil {
		return errors.New("failed updating additional specs")
	}

	itemspecsat := models.AdditionalSpecsAt{
		RefId:                  additionalspec.ID,
		AdditionalSpecsContent: additionalspec.AdditionalSpecsContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &itemspecsat); err != nil {
		return errors.New("failed creating itemspecsat")
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
