package services

import (
	"github.com/pierceperado/smpc/initializers"
	"gorm.io/gorm"
)

func DbGet(model interface{}, conditions map[string]interface{}) error {
	query := initializers.DB.Model(model)

	if conditions != nil {
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

// func Update(tx *gorm.DB,model,interface{}, conditions map[string]interface{}, updates map[string]interface{}) error {
// 	// Ensure that the model is passed as a pointer.
// 	query := tx.Model(model)

// 	// Apply conditions if provided
// 	if len(conditions) > 0 {
// 		query = query.Where(conditions)
// 	}

// 	// Apply updates
// 	if err := query.UpdateColumns(updates).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }

// func Delete(tx *gorm.DB,model interface{}, conditions map[string]interface{}) error {
// 	// Ensure that the model is passed as a pointer.
// 	query := initializers.DB.Model(model)

// 	// Apply conditions if provided
// 	if len(conditions) > 0 {
// 		query = query.Where(conditions)
// 	}

// 	// Perform the delete operation
// 	if err := query.Delete(model).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }
