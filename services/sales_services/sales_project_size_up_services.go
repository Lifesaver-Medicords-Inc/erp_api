package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// CreateProjectSizeUp - parentId is the owning content row's ContentID, assigned to the
// foreign key. Deliberately NOT to SizeUp.ID: that was the bug that broke the sibling
// Final Selection save (parent id written to the child's primary key, leaving the FK at
// zero so the preload never matched), so this mirrors the corrected shape.
func CreateProjectSizeUp(tx *gorm.DB, parentId uint, SizeUp models.SalesProjectSizeUp, at models.At) error {
	SizeUp.ID = 0
	SizeUp.SalesProjectContentID = parentId

	if err := services.DbInsert(tx, &SizeUp); err != nil {
		fmt.Println(err)
		return errors.New("failed creating project size up")
	}

	sizeupat := models.SalesProjectSizeUpAt{
		RefID:                    SizeUp.ID,
		SalesProjectSizeUpContent: SizeUp.SalesProjectSizeUpContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &sizeupat); err != nil {
		return errors.New("failed creating project size up audit row")
	}
	return nil
}

func GetSalesProjectSizeUp(SizeUp *[]models.SalesProjectSizeUp, conditions map[string]interface{}) error {
	if err := services.DbGet(SizeUp, conditions); err != nil {
		return errors.New("failed getting project size up")
	}
	return nil
}

func UpdateProjectSizeUp(tx *gorm.DB, SizeUp models.SalesProjectSizeUp, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &SizeUp, conditions); err != nil {
		return errors.New("failed updating project size up")
	}

	sizeupat := models.SalesProjectSizeUpAt{
		RefID:                    SizeUp.ID,
		SalesProjectSizeUpContent: SizeUp.SalesProjectSizeUpContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &sizeupat); err != nil {
		return errors.New("failed creating project size up audit row")
	}
	return nil
}

// DeleteProjectSizeUp - Size Up is a list the user adds to and removes from, so removal
// has to persist or a dropped candidate would reappear on reload (and keep constraining
// Final Selection). The audit row is written first so the deletion itself stays traceable
// in z_tbl_trans_sales_project_size_up_at.
func DeleteProjectSizeUp(tx *gorm.DB, SizeUp models.SalesProjectSizeUp, at models.At) error {
	sizeupat := models.SalesProjectSizeUpAt{
		RefID:                    SizeUp.ID,
		SalesProjectSizeUpContent: SizeUp.SalesProjectSizeUpContent,
		At:                       at,
	}

	if err := services.DbInsert(tx, &sizeupat); err != nil {
		return errors.New("failed creating project size up audit row")
	}

	if err := tx.Delete(&models.SalesProjectSizeUp{}, SizeUp.ID).Error; err != nil {
		return errors.New("failed deleting project size up")
	}
	return nil
}
