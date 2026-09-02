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

	// SalesProjectContentFinal and SalesProjectSizeUp are declared as GORM associations on
	// this struct, so tx.Create inserts them automatically while saving the parent - and the
	// explicit loops below then inserted them a SECOND time. That is why one save produced
	// duplicate finals. Detach them before inserting the parent and create them explicitly
	// afterwards: the explicit path is the one that zeroes the client-supplied id (avoiding
	// IDENTITY_INSERT collisions) and stamps the FK from the parent's freshly assigned
	// ContentID, which the auto-save cannot do because that id isn't known until this
	// insert returns.
	sizeUps := ProjectContent.SalesProjectSizeUp
	finals := ProjectContent.SalesProjectContentFinal
	ProjectContent.SalesProjectSizeUp = nil
	ProjectContent.SalesProjectContentFinal = nil

	fmt.Printf("DEBUG-MARKER-9f3a CreateProjectContent about to insert, ContentID=%d\n", ProjectContent.ContentID)

	if err := services.DbInsert(tx, &ProjectContent); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &ProjectContent)
		return errors.New("failed creating project content")
	}

	// Re-enabled. This was disabled ("Because of the foreign key I remove the manually
	// added") because it passed v.ID - the child's own id - as the parent key, which
	// combined with CreateProjectContentFinal assigning that to the child's primary key
	// made an FK violation unavoidable. Both halves are fixed now: pass the owning
	// content's ContentID, and that function sets SalesProjectContentID from it. Without
	// this loop the finals were loaded by GetSalesProjectContent's preload but never
	// written, so Final Selection always came back empty.
	for _, v := range finals {
		if err := CreateProjectContentFinal(tx, ProjectContent.ContentID, v, at); err != nil {
			return errors.New("failed creating project content finals")
		}
	}

	// Size Up (spec 5.1.4) - same child-of-content shape as the finals above.
	for _, v := range sizeUps {
		if err := CreateProjectSizeUp(tx, ProjectContent.ContentID, v, at); err != nil {
			return errors.New("failed creating project size up")
		}
	}

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
	if err := services.DbGetWithPreloads(ProjectContent, conditions, "SalesProjectContentFinal", "SalesProjectSizeUp"); err != nil {
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

	// Finals were handled on neither create (commented out) nor update (absent entirely),
	// so editing a saved project quote never persisted a Final Selection change either.
	// Upsert rather than replace: a row the client already knows the id of is updated in
	// place, a new one is created. That keeps existing ids stable for the audit trail in
	// z_tbl_trans_sales_project_content_final_at and avoids deleting rows outright.
	for _, v := range projectcontent.SalesProjectContentFinal {
		if v.ID > 0 {
			finalConditions := map[string]interface{}{"id": v.ID}
			if err := UpdateProjectContentFinal(tx, v, at, finalConditions); err != nil {
				return errors.New("failed updating project content finals")
			}
			continue
		}

		if err := CreateProjectContentFinal(tx, projectcontent.ContentID, v, at); err != nil {
			return errors.New("failed creating project content finals")
		}
	}

	// Size Up: upsert what came in, then remove any row this content still has in the DB
	// that the client didn't send back - that's how a candidate removed in the UI actually
	// disappears. Finals aren't pruned this way because the UI has no remove action for
	// them; Size Up does.
	keptSizeUpIds := map[uint]bool{}
	for _, v := range projectcontent.SalesProjectSizeUp {
		if v.ID > 0 {
			keptSizeUpIds[v.ID] = true
			sizeUpConditions := map[string]interface{}{"id": v.ID}
			if err := UpdateProjectSizeUp(tx, v, at, sizeUpConditions); err != nil {
				return errors.New("failed updating project size up")
			}
			continue
		}

		if err := CreateProjectSizeUp(tx, projectcontent.ContentID, v, at); err != nil {
			return errors.New("failed creating project size up")
		}
	}

	var existingSizeUps []models.SalesProjectSizeUp
	if err := tx.Where("sales_project_content_id = ?", projectcontent.ContentID).Find(&existingSizeUps).Error; err != nil {
		return errors.New("failed reading existing project size up")
	}

	for _, existing := range existingSizeUps {
		if keptSizeUpIds[existing.ID] {
			continue
		}
		if err := DeleteProjectSizeUp(tx, existing, at); err != nil {
			return errors.New("failed deleting project size up")
		}
	}

	return nil
}
