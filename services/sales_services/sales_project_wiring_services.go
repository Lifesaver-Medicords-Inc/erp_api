package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectWiring(tx *gorm.DB, parentId uint, projectwiring models.SalesProjectWiring, at models.At) error {
	projectwiring.BasedId = parentId

	if err := services.DbInsert(tx, &projectwiring); err != nil {
		return errors.New("failed creating project wirings")
	}

	projectwiringat := models.SalesProjectWiringAt{
		RefId:                     projectwiring.ID,
		SalesProjectWiringContent: projectwiring.SalesProjectWiringContent,
		At:                        at,
	}

	if err := services.DbInsert(tx, &projectwiringat); err != nil {
		return errors.New("failed creating wiring AT")
	}

	return nil
}

func GetProjectWiring(ProjectWiring *[]models.SalesProjectWiring, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectWiring, conditions); err != nil {
		return errors.New("failed getting project wiring")
	}
	return nil
}
