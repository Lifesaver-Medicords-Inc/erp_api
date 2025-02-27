package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectContent(tx *gorm.DB, parentId uint, ProjectContent models.SalesProjectContent, at models.At) error {

	ProjectContent.BasedId = parentId

	if err := services.DbInsert(tx, &ProjectContent); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &ProjectContent)
		return errors.New("failed creating project content")
	}

	projectcontentat := models.SalesProjectContentAt{
		RefID:                      ProjectContent.ID,
		SalesProjectContentContent: ProjectContent.SalesProjectContentContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}
	return nil
}

func GetSalesProjectContent(ProjectContent *[]models.SalesProjectContent, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectContent, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}

func UpdateProjectContent(tx *gorm.DB, projectcontent models.SalesProjectContent, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectcontent, conditions); err != nil {
		return errors.New("failed updating project content")
	}

	projectcontentat := models.SalesProjectContentAt{
		RefID:                      projectcontent.ID,
		SalesProjectContentContent: projectcontent.SalesProjectContentContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}

	return nil
}
