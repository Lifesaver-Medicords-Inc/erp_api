package sales_quotation_services

import (
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

func Get() ([]models.Brand, error) {

	var brand []models.Brand

	if err := initializers.DB.Find(&brand).Error; err != nil {
		return []models.Brand{}, err
	}

	return brand, nil
}

func Create(data *models.Brand) error {
	return nil
}

func Update(a int, b int) int {
	return a * b
}

func Delete(a int, b int) int {
	return a * b
}
