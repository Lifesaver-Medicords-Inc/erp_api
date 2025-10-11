package dispatching_services

import (
	"errors"
	"net/http"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
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
func (s *ItemReleaseService) CreateItemReleaseService(release *models.ItemReleaseModel) (*models.ItemReleaseModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return release, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Create(release).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, err
	}

	return release, http.StatusCreated, nil
}

// Update existing item release
func (s *ItemReleaseService) UpdateItemReleaseService(release *models.ItemReleaseModel, conditions map[string]interface{}) (*models.ItemReleaseModel, int, error) {
	var existing = &models.ItemReleaseModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return existing, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(&existing).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusNotFound, err
	}

	if err := tx.Model(&existing).Updates(release).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, err
	}

	return existing, http.StatusOK, nil
}

// Delete an item release
func (s *ItemReleaseService) DeleteItemReleaseService(conditions map[string]interface{}) (bool, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return false, 500, errors.New("failed to start DB transaction")
	}

	if tx.Error != nil {
		return false, http.StatusInternalServerError, tx.Error
	}

	if err := tx.Where(conditions).Delete(&models.ItemReleaseModel{}).Error; err != nil {
		tx.Rollback()
		return false, http.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return false, http.StatusInternalServerError, err
	}

	return true, http.StatusOK, nil
}
