package setup_services

import (
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetClasses() ([]models.Class, error) {
	var class []models.Class

	if err := services.DbGet(&class, nil); err != nil {
		return class, err
	}

	return class, nil
}
func GetClass(id int) (models.Class, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var class models.Class

	if err := services.DbGet(&class, conditions); err != nil {
		return class, err
	}

	return class, nil
}

func CreateClasses(tx *gorm.DB, data *models.Class) error {
	if err := services.DbInsert(tx, data); err != nil {
		return err
	}

	return nil
}
