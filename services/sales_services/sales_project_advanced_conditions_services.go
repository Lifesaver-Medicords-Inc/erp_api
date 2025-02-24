package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectAdvancedConditions(tx *gorm.DB, parentId uint, ProjectConditions models.SalesProjectAdvancedConditions, at models.At) error {
	ProjectConditions.BasedId = parentId

	if err := services.DbInsert(tx, &ProjectConditions); err != nil {
		return errors.New("failed creating project advanced conditions")
	}

	// projectconditionsat := models.SalesProjectAdvancedConditionsAt{
	// 	RefID:                          ProjectConditions.ID,
	// 	SalesProjectAdvancedConditions: ProjectConditions,
	// 	At:                             at,
	// }

	// if err := services.DbInsert(tx, &projectconditionsat); err != nil {
	// 	return errors.New("failed creating content child")
	// }
	return nil
}

func GetProjectAdvancedConditions(ProjectConditions *[]models.SalesProjectAdvancedConditions, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectConditions, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}
