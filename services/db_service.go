package services

import (
	"github.com/pierceperado/smpc/initializers"
	"gorm.io/gorm"
)

func DbGet(model interface{}, conditions map[string]interface{}) error {
	query := initializers.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Find(model).Error; err != nil {
		return err
	}

	return nil
}

func DbInsert(tx *gorm.DB, model interface{}) error {
	if err := tx.Create(model).Error; err != nil {
		return err
	}

	return nil
}

func DbUpdate(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	query := tx.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.UpdateColumns(model).Error; err != nil {
		return err
	}

	return nil
}

func DbDelete(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	query := initializers.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Delete(model).Error; err != nil {
		return err
	}

	return nil
}
