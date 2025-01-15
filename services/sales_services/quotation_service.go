package sales_services

import (
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

func GetQuotations() ([]models.Brand, error) {
	var brands []models.Brand

	if err := initializers.DB.Find(&brands).Error; err != nil {
		return brands, err
	}

	return brands, nil
}

func CreateQuotation(data *models.Brand) error {
	return nil
}

func UpdateQuotation(a int, b int) int {
	return a * b
}

func DeleteQuotation(a int, b int) int {
	return a * b
}
