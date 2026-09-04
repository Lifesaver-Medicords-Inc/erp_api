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
	// parentId is the owning content row's ContentID. This used to assign it to
	// ProjectContentFinal.ID - the child's OWN primary key - which both forced an
	// IDENTITY_INSERT of an id that may already exist and left the real foreign key
	// (SalesProjectContentID, per the gorm tag on SalesProjectContent) at zero, so the
	// row could never be preloaded back onto its parent. That FK failure is why the
	// caller in CreateProjectContent was commented out rather than fixed, and why
	// Final Selection never came back on the UI. Same reasoning as CreateProjectContent's
	// own "always let the DB assign a fresh id here" fix.
	ProjectContentFinal.ID = 0
	ProjectContentFinal.SalesProjectContentID = parentId

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

// DeleteProjectContentFinal mirrors DeleteProjectSizeUp - audit row first, then the
// delete, so the removal is recoverable from
// z_tbl_trans_sales_project_content_final_at. Added 2026-09-03 when finals started
// being pruned like size-ups (syncContentChildren); before that nothing ever removed
// a final, which is why this had no delete path at all.
func DeleteProjectContentFinal(tx *gorm.DB, ProjectContentFinal models.SalesProjectContentFinal, at models.At) error {
	finalat := models.SalesProjectContentFinalAt{
		RefID:                           ProjectContentFinal.ID,
		SalesProjectContentFinalContent: ProjectContentFinal.SalesProjectContentFinalContent,
		At:                              at,
	}

	if err := services.DbInsert(tx, &finalat); err != nil {
		return errors.New("failed creating project content final audit row")
	}

	if err := tx.Delete(&models.SalesProjectContentFinal{}, ProjectContentFinal.ID).Error; err != nil {
		return errors.New("failed deleting project content final")
	}

	return nil
}
