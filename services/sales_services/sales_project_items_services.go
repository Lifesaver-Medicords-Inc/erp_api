package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItems(tx *gorm.DB, parentId uint, ProjectItems models.SalesProjectItems, at models.At) error {
	ProjectItems.BasedId = parentId

	if err := services.DbInsert(tx, &ProjectItems); err != nil {
		return errors.New("failed creating project advanced conditions")
	}

	// projectconditionsat := models.SalesProjectItemsAt{
	// 	RefID:             ProjectItems.ID,
	// 	SalesProjectItems: ProjectItems,
	// 	At:                at,
	// }

	// if err := services.DbInsert(tx, &projectconditionsat); err != nil {
	// 	return errors.New("failed creating content child")
	// }
	return nil
}

func GetProjectItems(ProjectItems *[]models.SalesProjectItems, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectItems, conditions); err != nil {
		return errors.New("failed getting project items")
	}
	return nil
}
