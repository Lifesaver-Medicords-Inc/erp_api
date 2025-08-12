package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type ItemSpecsContent struct {
	BasedId  uint    `json:"based_id"`
	Template string  `json:"template"`
	Field    []Field `json:"fields"`
}

func GetItemSpecs(itemspecs *[]models.ItemSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(itemspecs, conditions); err != nil {
		return errors.New("failed getting itemspecs")
	}
	return nil
}

func GetItemSpec(itemspec *models.ItemSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(itemspec, conditions); err != nil {
		return errors.New("failed getting itemspec")
	}

	return nil
}

func CreateItemSpec(tx *gorm.DB, basedId uint, itemSpecs ItemSpecs, at models.At) error {
	for _, field := range itemSpecs.Fields {
		content := models.ItemSpecsContent{
			BasedId:            basedId,
			Template:           itemSpecs.Template,
			Title:              field.Title,
			Value:              field.Value,
			ManufacturerOrigin: itemSpecs.ManufacturerOrigin,
		}

		itemspecs := models.ItemSpecs{ItemSpecsContent: content}
		if err := services.DbInsert(tx, &itemspecs); err != nil {
			return errors.New("failed creating itemspecs")
		}

		itemspecsat := models.ItemSpecsAt{
			RefId:            itemspecs.ID,
			ItemSpecsContent: content,
			At:               at,
		}

		if err := services.DbInsert(tx, &itemspecsat); err != nil {
			return errors.New("failed creating itemspecsat")
		}
	}

	return nil
}

func UpdateItemSpec(tx *gorm.DB, basedId uint, itemSpecs ItemSpecs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &models.ItemSpecs{}, conditions); err != nil {
		return errors.New("failed deleting existing itemspecs")
	}

	for _, field := range itemSpecs.Fields {
		content := models.ItemSpecsContent{
			BasedId:            basedId,
			Template:           itemSpecs.Template,
			Title:              field.Title,
			Value:              field.Value,
			ManufacturerOrigin: itemSpecs.ManufacturerOrigin,
		}

		itemspecs := models.ItemSpecs{ItemSpecsContent: content}
		if err := services.DbInsert(tx, &itemspecs); err != nil {
			return errors.New("failed creating updated itemspecs")
		}

		itemspecsat := models.ItemSpecsAt{
			RefId:            itemspecs.ID,
			ItemSpecsContent: content,
			At:               at,
		}

		if err := services.DbInsert(tx, &itemspecsat); err != nil {
			return errors.New("failed creating updated itemspecsat")
		}
	}

	return nil
}

func DeleteItemSpecs(tx *gorm.DB, itemspecs models.ItemSpecs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &itemspecs, conditions); err != nil {
		return errors.New("failed deleting itemspecs")
	}

	itemspecsat := models.ItemSpecsAt{
		RefId:            itemspecs.ID,
		ItemSpecsContent: itemspecs.ItemSpecsContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &itemspecsat); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	return nil
}
