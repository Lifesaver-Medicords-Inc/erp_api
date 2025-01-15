package setup_services

import (
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetBrands() ([]models.Brand, error) {
	var brands []models.Brand

	if err := services.DbGet(&brands, nil); err != nil {
		return brands, err
	}

	return brands, nil
}

func GetBrand(id int) (models.Brand, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var brand models.Brand

	if err := services.DbGet(&brand, conditions); err != nil {
		return brand, err
	}

	return brand, nil
}

func CreateBrand(tx *gorm.DB, data *models.Brand) error {
	if err := services.DbInsert(tx, data); err != nil {
		return err
	}

	return nil
}

// func Update(a int, b int) int {
// 	return a * b
// }

// func Delete(a int, b int) int {
// 	return a * b
// }
