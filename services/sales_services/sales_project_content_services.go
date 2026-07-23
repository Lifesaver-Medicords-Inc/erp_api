package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectContent(tx *gorm.DB, parentId uint, ProjectContent models.SalesProjectContent, at models.At) error {
	ProjectContent.BasedId = parentId

	// Same fix as CreateProjectItemSet's ItemSetID = 0: ContentID is client-controlled on the
	// wire (an existing tab's content reports its own id back to the server elsewhere), but
	// this function only ever creates a brand new content row. Trusting whatever id the client
	// sent forced GORM to SET IDENTITY_INSERT ON ... VALUES(..., <that id>), which collided
	// with a PRIMARY KEY violation whenever that id already existed - e.g. a leftover/stale
	// content_id carried over in the UI from a previously loaded tab. Always let the DB assign
	// a fresh id here.
	ProjectContent.ContentID = 0

	fmt.Printf("DEBUG-MARKER-9f3a CreateProjectContent about to insert, ContentID=%d\n", ProjectContent.ContentID)

	if err := services.DbInsert(tx, &ProjectContent); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &ProjectContent)
		return errors.New("failed creating project content")
	}

	// Because of the foreign key I remove the manually added
	// for _, v := range ProjectContent.SalesProjectContentFinal {
	// 	if err := CreateProjectContentFinal(tx, v.ID, v, at); err != nil {
	// 		return errors.New("failed creating project content finals")
	// 	}
	// }

	projectcontentat := models.SalesProjectContentAt{
		RefID:                      ProjectContent.ContentID,
		SalesProjectContentContent: ProjectContent.SalesProjectContentContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}
	return nil
}

func GetSalesProjectContent(ProjectContent *[]models.SalesProjectContent, conditions map[string]interface{}) error {
	if err := services.DbGetWithPreloads(ProjectContent, conditions, "SalesProjectContentFinal"); err != nil {
		return errors.New("failed getting project content")
	}
	return nil
}

func GetSalesProjectCont(id int) (models.SalesProjectContent, int, error) {
	conditions := map[string]interface{}{
		"based_id": id,
	}

	var projectcontent models.SalesProjectContent

	if err := services.DbGet(&projectcontent, conditions); err != nil {
		return projectcontent, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return projectcontent, 0, nil
}

func UpdateProjectContent(tx *gorm.DB, projectcontent models.SalesProjectContent, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &projectcontent, conditions); err != nil {
		return errors.New("failed updating project content")
	}

	projectcontentat := models.SalesProjectContentAt{
		RefID:                      projectcontent.ContentID,
		SalesProjectContentContent: projectcontent.SalesProjectContentContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &projectcontentat); err != nil {
		return errors.New("failed creating project content")
	}

	return nil
}
