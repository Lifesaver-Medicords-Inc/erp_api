package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectWiring(tx *gorm.DB, parentId uint, projectwiring models.SalesProjectWiring, at models.At) error {
	projectwiring.BasedId = parentId

	// Same fix as CreateProjectItemSet's ItemSetID = 0 - ID is client-controlled on the wire,
	// but this function only ever creates a brand new row. Always let the DB assign a fresh id
	// instead of trusting whatever (possibly stale/leftover) id the client sent.
	projectwiring.ID = 0

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

func UpdateProjectWiring(tx *gorm.DB, projectWirings models.SalesProjectWiring, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectWirings, conditions); err != nil {
		return errors.New("failed updating project wirings")
	}

	projectwiringsat := models.SalesProjectWiringAt{
		RefId:                     projectWirings.ID,
		SalesProjectWiringContent: projectWirings.SalesProjectWiringContent,
		At:                        at,
	}

	if err := services.DbInsert(tx, &projectwiringsat); err != nil {
		return errors.New("failed creating project wirings at")
	}

	return nil
}
