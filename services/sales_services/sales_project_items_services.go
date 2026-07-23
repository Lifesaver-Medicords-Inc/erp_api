package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItems(tx *gorm.DB, parentId uint, projectitems models.SalesProjectItems, images []models.SalesQuotationSelectedImage, at models.At) error {
	projectitems.BasedId = parentId

	// Same fix as CreateProjectItemSet's ItemSetID = 0 - ItemsID is client-controlled on the
	// wire, but this function only ever creates a brand new item row. Always let the DB assign
	// a fresh id instead of trusting whatever (possibly stale/leftover) id the client sent.
	projectitems.ItemsID = 0

	if err := services.DbInsert(tx, &projectitems); err != nil {
		return errors.New("failed creating project items")
	}

	projectconditionsat := models.SalesProjectItemsAt{
		RefID:                    projectitems.ItemsID,
		SalesProjectItemsContent: projectitems.SalesProjectItemsContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &projectconditionsat); err != nil {
		return errors.New("failed creating content child")
	}

	// Reuses the same table/service Quick Quote items already use for selected images -
	// the column is still named quotation_quick_id, but it's just a plain FK-style int
	// column (no DB-level constraint), and SalesProjectItems.ItemsID is a valid ID to
	// key it by here. DbInsert sets ItemsID on the struct above, so it's available now.
	if len(images) > 0 {
		if err := CreateSalesQuotationSelectedImages(tx, projectitems.ItemsID, images, at); err != nil {
			return err
		}
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

	projectitemsat := models.SalesProjectItemsAt{
		RefID:                    projectitems.ItemsID,
		SalesProjectItemsContent: projectitems.SalesProjectItemsContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &projectitemsat); err != nil {
		return errors.New("failed creating content child")
	}

	return nil
}
