package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItems(tx *gorm.DB, parentId uint, projectitems models.SalesProjectItems, at models.At) error {
	projectitems.BasedId = parentId

	if err := services.DbInsert(tx, &projectitems); err != nil {
		return errors.New("failed creating project advanced conditions")
	}

	projectconditionsat := models.SalesProjectItemsAt{
		RefID:                    projectitems.ItemsID,
		SalesProjectItemsContent: projectitems.SalesProjectItemsContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &projectconditionsat); err != nil {
		return errors.New("failed creating content child")
	}

	return nil
}

func GetProjectItems(ProjectItems *[]models.SalesProjectItems, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectItems, conditions); err != nil {
		return errors.New("failed getting project items")
	}
	return nil
}

func UpdateProjectItems(tx *gorm.DB, projectitems models.SalesProjectItems, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectitems, conditions); err != nil {
		return errors.New("failed updating project items")
	}

	projectconditionsat := models.SalesProjectItemsAt{
		RefID:                    projectitems.ItemsID,
		SalesProjectItemsContent: projectitems.SalesProjectItemsContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &projectconditionsat); err != nil {
		return errors.New("failed creating content child")
	}

	return nil
}
