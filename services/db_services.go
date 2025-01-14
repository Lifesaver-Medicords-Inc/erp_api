package db_services

import (
	"github.com/pierceperado/smpc/initializers"
	"gorm.io/gorm"
)

func Get(model interface{}, conditions map[string]interface{}) error {
	// Ensure that we pass a pointer to `model`, so GORM can modify it.
	query := initializers.DB.Model(model)

	// Apply conditions if provided (conditions is nil or empty)
	if conditions != nil && len(conditions) > 0 {
		query = query.Where(conditions)
	}

	// Execute the query and retrieve the data
	if err := query.Find(model).Error; err != nil {
		return err
	}
	return nil
}

func Insert(tx *gorm.DB, model interface{}) error {
	// Ensure that the model is passed as a pointer.
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
