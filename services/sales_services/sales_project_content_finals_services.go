package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func  CreateProjectContentFinal(tx *gorm.DB, parentId uint, ProjectContentFinal models.SalesProjectContentFinal, at models.At) error {
	ProjectContentFinal.ID = parentId

	if err := services.DbInsert(tx, &ProjectContentFinal); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &ProjectContentFinal)
		return errors.New("failed creating project content final")
	}

	projectcontentat := models.SalesProjectContentFinalAt{
		RefID:                           ProjectContentFinal.ID,
		SalesProjectContentFinalContent: ProjectContentFinal.SalesProjectContentFinalContent,
		At:                              at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}
	return nil
}

func GetSalesProjectContentFinal(ProjectContentFinal *[]models.SalesProjectContentFinal, conditions map[string]interface{}) error {
	if err := services.DbGet(ProjectContentFinal, conditions); err != nil {
		return errors.New("failed getting multipliers")
	}
	return nil
}

func GetSalesProjectContFinal(id int) (models.SalesProjectContentFinal, int, error) {
	conditions := map[string]interface{}{
		"base_project_content_id": id,
	}

	var projectcontentfinal models.SalesProjectContentFinal

	if err := services.DbGet(&projectcontentfinal, conditions); err != nil {
		return projectcontentfinal, fiber.StatusInternalServerError, errors.New("failed getting final")
	}

	return projectcontentfinal, 0, nil
}

func UpdateProjectContentFinal(tx *gorm.DB, projectcontentfinal models.SalesProjectContentFinal, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectcontentfinal, conditions); err != nil {
		return errors.New("failed updating project content")
	}

	projectcontentat := models.SalesProjectContentFinalAt{
		RefID:                           projectcontentfinal.ID,
		SalesProjectContentFinalContent: projectcontentfinal.SalesProjectContentFinalContent,
		At:                              at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}

	return nil
}
