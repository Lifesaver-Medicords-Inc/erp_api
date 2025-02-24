package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectContentChild(tx *gorm.DB, parentId uint, ProjectContentChild models.SalesProjectContentChild, at models.At) error {
	ProjectContentChild.BasedID = parentId

	if err := services.DbInsert(tx, &ProjectContentChild); err != nil {
		return errors.New("failed creating project content child")
	}

	// projectcontentchildat := models.SalesProjectContentChildAt{
	// 	RefId:                    ProjectContentChild.ID,
	// 	SalesProjectContentChild: ProjectContentChild,
	// 	At:                       at,
	// }

	// if err := services.DbInsert(tx, &projectcontentchildat); err != nil {
	// 	return errors.New("failed creating content child")
	// }
	return nil
}

func GetSalesProjectContentChild(ProjectContentChild *[]models.SalesProjectContentChild, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectContentChild, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}
