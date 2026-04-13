package dispatching_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

type ItemReleaseService struct{}

func NewItemReleaseService() *ItemReleaseService {
	return &ItemReleaseService{}
}

// Get all item releases with optional conditions
func (s *ItemReleaseService) GetItemReleasesService(conditions map[string]interface{}) ([]models.ItemRelease, int, error) {
	var releases = []models.ItemRelease{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return releases, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbGetRel(&releases, conditions, "ItemReleaseDetails"); err != nil {
		return releases, fiber.StatusInternalServerError, err
	}

	return releases, fiber.StatusOK, nil
}

// Get a single item release
func (s *ItemReleaseService) GetItemReleaseService(conditions map[string]interface{}) (*models.ItemRelease, int, error) {
	var release = &models.ItemRelease{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return release, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbGetRel(release, conditions, "ItemReleaseDetails"); err != nil {
		return release, fiber.StatusNotFound, err
	}
	return release, fiber.StatusOK, nil
}

// Create a new item release
func (s *ItemReleaseService) CreateItemReleaseService(release *models.ItemRelease, at models.At) (*models.ItemRelease, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return release, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	nextDocNo, err := utils.NextDocNo(tx, new(models.ItemRelease), "doc_no")
	if err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	release.DocNo = nextDocNo

	if err := services.DbInsert(tx, release); err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed creating item release")
	}

	atdata := models.ItemReleaseAt{
		RefId: release.ID,
		At:    at,
	}

	if err := tx.Create(&atdata).Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed creating releaseat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateIRCaches()

	return release, fiber.StatusCreated, nil
}

func (s *ItemReleaseService) UpdateItemReleaseService(release *models.ItemRelease, conditions map[string]interface{}, at models.At) (*models.ItemRelease, int, error) {

	// Start transaction
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, release, conditions); err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed updating item release")
	}

	// Insert audit record
	atdata := models.ItemReleaseAt{
		RefId: release.ID,
		At:    at,
	}
	if err := tx.Create(&atdata).Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed creating release audit")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateIRCaches()

	// Return updated record
	return release, fiber.StatusOK, nil
}

// Delete an item release
func (s *ItemReleaseService) DeleteItemReleaseService(conditions map[string]interface{}, at models.At) (*models.ItemRelease, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &models.ItemRelease{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	release, status, err := s.GetItemReleaseService(conditions)
	if err != nil {
		return release, status, errors.New("calendar release not found")
	}

	if err := services.DbDelete(tx, &release, conditions); err != nil {
		return release, fiber.StatusInternalServerError, errors.New("failed deleting calendar release")
	}

	atdata := models.ItemReleaseAt{RefId: release.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed creating release audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateIRCaches()

	return release, fiber.StatusOK, nil
}

func (s *ItemReleaseService) GetSalesOrderDetails(conditions map[string]interface{}) ([]models.SalesOrderItemReleaseView, int, error) {
	var releases []models.SalesOrderItemReleaseView

	if err := services.DbGet(&releases, conditions); err != nil {
		return releases, fiber.StatusInternalServerError, errors.New("failed getting so with approved ir")
	}

	return releases, fiber.StatusOK, nil
}

func (s *ItemReleaseService) GetItemStockAndLocation(itemId uint) ([]inventory_models.ItemStockAndLocationView, int, error) {
	conditions := map[string]interface{}{
		"ItemId": itemId,
	}

	var response []inventory_models.ItemStockAndLocationView

	// Call stored procedure
	if err := services.DbRaw(&response, "sp_GetItemStockAndLocation", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting it stock and locations")
	}

	return response, fiber.StatusOK, nil
}

func InvalidateIRCaches() {
	cacheKeys := []interface{}{
		models.SalesOrderItemReleaseView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
