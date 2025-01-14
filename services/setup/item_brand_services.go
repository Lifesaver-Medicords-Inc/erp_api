package item_brand_services

import (
	"github.com/pierceperado/smpc/models"
	db_services "github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func Get() ([]models.Brand, error) {

	var brand []models.Brand

	if err := db_services.Get(&brand, nil); err != nil {
		return []models.Brand{}, err
	}

	return brand, nil
}

func GetById(id int) ([]models.Brand, error) {

	conditions := map[string]interface{}{
		"id": id,
	}

	var brand []models.Brand

	if err := db_services.Get(&brand, conditions); err != nil {
		return []models.Brand{}, err
	}

	return brand, nil
}

func Create(tx *gorm.DB, data *models.Brand) (models.Brand, error) {

	var brand models.Brand

	if err := db_services.Insert(tx, &brand); err != nil {
		return models.Brand{}, err
	}

	return brand, nil
}

// func Update(a int, b int) int {
// 	return a * b
// }

// func Delete(a int, b int) int {
// 	return a * b
// }
