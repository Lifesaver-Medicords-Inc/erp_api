package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectAdvancedConditions(tx *gorm.DB, parentId uint, ProjectConditions models.SalesProjectAdvancedConditions, at models.At) error {
	ProjectConditions.BasedId = parentId

	// Same fix as CreateProjectItemSet's ItemSetID = 0 - ConditionsID is client-controlled on
	// the wire, but this function only ever creates a brand new row. Always let the DB assign
	// a fresh id instead of trusting whatever (possibly stale/leftover) id the client sent.
	ProjectConditions.ConditionsID = 0

	if err := services.DbInsert(tx, &ProjectConditions); err != nil {
		return errors.New("failed creating project advanced conditions")
	}

	projectconditionsat := models.SalesProjectAdvancedConditionsAt{
		RefID:                                 ProjectConditions.ConditionsID,
		SalesProjectAdvancedConditionsContent: ProjectConditions.SalesProjectAdvancedConditionsContent,
		At:                                    at,
	}

	if err := services.DbInsert(tx, &projectconditionsat); err != nil {
		return errors.New("failed creating content child")
	}
	return nil
}

func GetProjectAdvancedConditions(ProjectConditions *[]models.SalesProjectAdvancedConditions, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectConditions, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}

func UpdateProjectAdvancedConditions(tx *gorm.DB, projectconditions *models.SalesProjectAdvancedConditions, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, projectconditions, conditions); err != nil {
		return errors.New("failed updating project advanced conditions")
	}

	projectconditionsat := models.SalesProjectAdvancedConditionsAt{
		RefID:                                 projectconditions.ConditionsID,
		SalesProjectAdvancedConditionsContent: projectconditions.SalesProjectAdvancedConditionsContent,
		At:                                    at,
	}

	if err := services.DbInsert(tx, &projectconditionsat); err != nil {
		return errors.New("failed creating project conditions")
	}

	return nil
}
