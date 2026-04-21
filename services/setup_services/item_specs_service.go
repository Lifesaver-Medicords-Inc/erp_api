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
	if err := services.DbGetWithPreloads(itemspecs, conditions, "ItemSpecsTemplate"); err != nil {
		return errors.New("failed getting itemspecs")
	}
	return nil
}

func GetItemSpec(itemspec *models.ItemSpecs, conditions map[string]interface{}) error {
	if err := services.DbGetWithPreloads(itemspec, conditions, "ItemSpecsTemplate"); err != nil {
		return errors.New("failed getting itemspec")
	}

	return nil
}

func CreateItemSpec(tx *gorm.DB, basedId uint, itemSpecs models.ItemSpecs, at models.At) error {
	itemspecs := models.ItemSpecs{ItemSpecsContent: models.ItemSpecsContent{
		BasedId:            basedId,
		Template:           itemSpecs.Template,
		ManufacturerOrigin: itemSpecs.ManufacturerOrigin,
		Fla_1:              itemSpecs.Fla_1,
		Fla_2:              itemSpecs.Fla_2,
		Volt_1:             itemSpecs.Volt_1,
		Volt_2:             itemSpecs.Volt_2,
	}}

	if err := services.DbInsert(tx, &itemspecs); err != nil {
		return errors.New("failed creating itemspecs")
	}

	if err := services.DbInsert(tx, &models.ItemSpecsAt{RefId: itemspecs.ID, ItemSpecsContent: itemspecs.ItemSpecsContent, At: at}); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	for _, field := range itemSpecs.ItemSpecsTemplate {
		t := models.ItemSpecsTemplate{ItemSpecsTemplateContent: models.ItemSpecsTemplateContent{BasedId: itemspecs.ID, Title: field.Title, Value: field.Value}}

		if err := services.DbInsert(tx, &t); err != nil {
			return errors.New("failed creating itemspecs template")
		}

		if err := services.DbInsert(tx, &models.ItemSpecsTemplateAt{RefId: t.ID, ItemSpecsTemplateContent: t.ItemSpecsTemplateContent, At: at}); err != nil {
			return errors.New("failed creating itemspecs template at")
		}
	}

	return nil
}

func UpdateItemSpec(tx *gorm.DB, basedId uint, itemSpecs models.ItemSpecs, at models.At, conditions map[string]interface{}) error {
	var existing models.ItemSpecs

	err := services.DbGet(&existing, conditions)

	// If no existing ItemSpecs record, create it from scratch
	if err != nil || existing.ID == 0 {
		return CreateItemSpec(tx, basedId, itemSpecs, at)
	}

	// Update existing ItemSpecs fields
	existing.Template = itemSpecs.Template
	existing.ManufacturerOrigin = itemSpecs.ManufacturerOrigin
	existing.Fla_1 = itemSpecs.Fla_1
	existing.Fla_2 = itemSpecs.Fla_2
	existing.Volt_1 = itemSpecs.Volt_1
	existing.Volt_2 = itemSpecs.Volt_2

	if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
		return errors.New("failed updating itemspecs")
	}

	if err := services.DbInsert(tx, &models.ItemSpecsAt{
		RefId:            existing.ID,
		ItemSpecsContent: existing.ItemSpecsContent,
		At:               at,
	}); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	// Get all existing templates for this ItemSpecs
	var existingTemplates []models.ItemSpecsTemplate
	if err := services.DbGet(&existingTemplates, map[string]interface{}{"based_id": existing.ID}); err != nil {
		return errors.New("failed getting existing templates")
	}

	// Map existing templates by ID for quick lookup
	existingMap := make(map[uint]models.ItemSpecsTemplate)
	for _, t := range existingTemplates {
		existingMap[t.ID] = t
	}

	// Track incoming IDs to detect deletions
	incomingIds := make(map[uint]bool)

	for _, field := range itemSpecs.ItemSpecsTemplate {
		if field.ID == 0 {
			// ✅ NEW: insert with correct existing.ID as BasedId
			t := models.ItemSpecsTemplate{
				ItemSpecsTemplateContent: models.ItemSpecsTemplateContent{
					BasedId: existing.ID,
					Title:   field.Title,
					Value:   field.Value,
				},
			}
			if err := services.DbInsert(tx, &t); err != nil {
				return errors.New("failed creating itemspecs template")
			}
			if err := services.DbInsert(tx, &models.ItemSpecsTemplateAt{
				RefId:                    t.ID,
				ItemSpecsTemplateContent: t.ItemSpecsTemplateContent,
				At:                       at,
			}); err != nil {
				return errors.New("failed creating itemspecs template at")
			}
		} else {
			// ✅ UPDATE: only if title or value changed
			incomingIds[field.ID] = true
			existingT, exists := existingMap[field.ID]
			if exists && (existingT.Title != field.Title || existingT.Value != field.Value) {
				existingT.Title = field.Title
				existingT.Value = field.Value
				if err := services.DbUpdate(tx, &existingT, map[string]interface{}{"id": existingT.ID}); err != nil {
					return errors.New("failed updating itemspecs template")
				}
				if err := services.DbInsert(tx, &models.ItemSpecsTemplateAt{
					RefId:                    existingT.ID,
					ItemSpecsTemplateContent: existingT.ItemSpecsTemplateContent,
					At:                       at,
				}); err != nil {
					return errors.New("failed creating itemspecs template at")
				}
			}
		}
	}

	// ✅ DELETE: any existing template not in incoming list
	for _, t := range existingTemplates {
		if !incomingIds[t.ID] {
			// Log to audit table before deleting
			if err := services.DbInsert(tx, &models.ItemSpecsTemplateAt{
				RefId:                    t.ID,
				ItemSpecsTemplateContent: t.ItemSpecsTemplateContent,
				At:                       at,
			}); err != nil {
				return errors.New("failed creating itemspecs template at")
			}
			if err := services.DbDelete(tx, &models.ItemSpecsTemplate{}, map[string]interface{}{"id": t.ID}); err != nil {
				return errors.New("failed deleting itemspecs template")
			}
		}
	}

	return nil
}
func DeleteItemSpec(tx *gorm.DB, conditions map[string]interface{}, at models.At) error {
	var itemspecs models.ItemSpecs
	if err := services.DbGet(&itemspecs, conditions); err != nil {
		return errors.New("failed getting itemspecs")
	}

	var templates []models.ItemSpecsTemplate
	if err := services.DbGet(&templates, map[string]interface{}{"based_id": itemspecs.ID}); err != nil {
		return errors.New("failed getting itemspecs templates")
	}

	for _, t := range templates {
		if err := services.DbInsert(tx, &models.ItemSpecsTemplateAt{RefId: t.ID, ItemSpecsTemplateContent: t.ItemSpecsTemplateContent, At: at}); err != nil {
			return errors.New("failed creating itemspecs template at")
		}
		if err := services.DbDelete(tx, &models.ItemSpecsTemplate{}, map[string]interface{}{"id": t.ID}); err != nil {
			return errors.New("failed deleting itemspecs template")
		}
	}

	if err := services.DbInsert(tx, &models.ItemSpecsAt{RefId: itemspecs.ID, ItemSpecsContent: itemspecs.ItemSpecsContent, At: at}); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	if err := services.DbDelete(tx, &models.ItemSpecs{}, conditions); err != nil {
		return errors.New("failed deleting itemspecs")
	}

	return nil
}
