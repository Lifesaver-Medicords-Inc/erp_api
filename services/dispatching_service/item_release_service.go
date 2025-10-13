package dispatching_services

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type ItemReleaseService struct{}

func NewItemReleaseService() *ItemReleaseService {
	return &ItemReleaseService{}
}

// Get all item releases with optional conditions
func (s *ItemReleaseService) GetItemReleasesService(conditions map[string]interface{}) ([]models.ItemReleaseModel, int, error) {
	var releases = []models.ItemReleaseModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return releases, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("ReleaseItems").Where(conditions).Find(&releases)

	if query.Error != nil {
		return nil, http.StatusInternalServerError, tx.Error
	}
	return releases, http.StatusOK, nil
}

// Get a single item release
func (s *ItemReleaseService) GetItemReleaseService(conditions map[string]interface{}) (*models.ItemReleaseModel, int, error) {
	var release = &models.ItemReleaseModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return release, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("ReleaseItems").Where(conditions).First(&release)

	if query.Error != nil {
		return nil, http.StatusNotFound, tx.Error
	}
	return release, http.StatusOK, nil
}

// Create a new item release
func (s *ItemReleaseService) CreateItemReleaseService(release *models.ItemReleaseModel, at models.At) (*models.ItemReleaseModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return release, 500, errors.New("failed to start DB transaction")
	}
	if err := services.DbInsert(tx, &release); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating release")
		}
		tx.Rollback()
		return release, 500, err
	}

	atdata := models.ItemReleaseAt{RefId: release.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed creating releaseat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed to commit transaction")
	}

	return release, 200, nil
}

// Update existing item release
func (s *ItemReleaseService) UpdateItemReleaseService(release *models.ItemReleaseModel, conditions map[string]interface{}, at models.At) (*models.ItemReleaseModel, int, error) {
	var existing = &models.ItemReleaseModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return existing, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.First(&release, conditions).Error; err != nil {
		return nil, http.StatusNotFound, err
	}

	if err := services.DbUpdate(tx, &release, conditions); err != nil {
		return release, 500, errors.New("failed updating release")
	}

	atdata := models.ItemReleaseAt{RefId: release.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed creating releaseat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed to commit transaction")
	}
	return release, 200, nil
}

// Delete an item release
func (s *ItemReleaseService) DeleteItemReleaseService(conditions map[string]interface{}, at models.At) (*models.ItemReleaseModel, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &models.ItemReleaseModel{}, 500, errors.New("failed to start DB transaction")
	}

	release, status, err := s.GetItemReleaseService(conditions)
	if err != nil {
		return release, status, errors.New("calendar release not found")
	}

	if err := services.DbDelete(tx, &release, conditions); err != nil {
		return release, 500, errors.New("failed deleting calendar release")
	}

	atdata := models.ItemReleaseAt{RefId: release.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed creating release audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed to commit transaction")
	}

	return release, 200, nil
}
