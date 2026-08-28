package dispatching_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type ItemReleaseService struct {
	stockService *item_stock_services.ItemStockService
}

func NewItemReleaseService() *ItemReleaseService {
	return &ItemReleaseService{
		stockService: item_stock_services.NewItemStockService(),
	}
}

// Get all item releases with optional conditions
func (s *ItemReleaseService) GetItemReleasesService(conditions map[string]interface{}) ([]models.ItemRelease, int, error) {
	var releases = []models.ItemRelease{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return releases, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbGetWithPreloads(&releases, conditions, "ItemReleaseDetails"); err != nil {
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

	if err := services.DbGetWithPreloads(release, conditions, "ItemReleaseDetails"); err != nil {
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

	if release.ItemReleaseDetails != nil {
		for i := range *release.ItemReleaseDetails {
			detail := &(*release.ItemReleaseDetails)[i]

			if err := services.RecomputeSoItemStatus(tx, detail.SalesOrderDetailsID); err != nil {
				tx.Rollback()
				return release, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
			}

			// Actually move stock - previously CreateItemReleaseService only ever wrote
			// the IREL document and its detail rows; nothing here ever called
			// DeductStockWithTx for any bin. See ApplyItemReleaseLocations' own comment.
			if err := s.ApplyItemReleaseLocations(tx, release, detail, at); err != nil {
				tx.Rollback()
				return release, fiber.StatusUnprocessableEntity, err
			}
		}
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

	// DbUpdate above only issues UPDATE on the item release's own columns; GORM's
	// UpdateColumns does not cascade to the has-many ItemReleaseDetails association,
	// so edited detail rows (e.g. released_qty from Pick Activity) must be saved here.
	if release.ItemReleaseDetails != nil {
		// Snapshot each detail's currently-persisted released_qty before DbUpdateDetails
		// overwrites it below - a line's bin allocation is only ever touched when its
		// released_qty actually changes (a plain Forward/Cancel Request resubmits the
		// same details untouched and must not re-deduct or restore anything). Missing
		// from the map (brand-new line, or one that fails to load) reads back as the
		// zero value, which correctly treats it as "changed" whenever ReleasedQty > 0.
		previousQty := map[uint]uint{}
		for _, detail := range *release.ItemReleaseDetails {
			if detail.ID == 0 {
				continue
			}
			var existing models.ItemReleaseDetails
			if err := tx.Where("id = ?", detail.ID).First(&existing).Error; err == nil {
				previousQty[detail.ID] = existing.ReleasedQty
			}
		}

		err := services.DbUpdateDetails(tx, *release.ItemReleaseDetails, func(d *models.ItemReleaseDetails) {
			d.ItemReleaseID = release.ID
		})
		if err != nil {
			tx.Rollback()
			return release, fiber.StatusInternalServerError, errors.New("failed saving item release details")
		}

		for i := range *release.ItemReleaseDetails {
			detail := &(*release.ItemReleaseDetails)[i]

			if err := services.RecomputeSoItemStatus(tx, detail.SalesOrderDetailsID); err != nil {
				tx.Rollback()
				return release, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
			}

			if previousQty[detail.ID] == detail.ReleasedQty {
				continue
			}

			// Released qty for this line is actually changing (including a fresh line
			// created by this same edit). Restore whatever it previously deducted, if
			// anything, then reapply the freshly submitted bin allocation wholesale -
			// see ApplyItemReleaseLocations/RestoreItemReleaseLocations doc comments for
			// why this restores-and-redoes rather than diffing bin-by-bin.
			remarks := fmt.Sprintf("Item Release #%d (edited)", release.DocNo)
			if err := s.RestoreItemReleaseLocations(tx, detail.ID, remarks, at); err != nil {
				tx.Rollback()
				return release, fiber.StatusInternalServerError, err
			}

			if err := s.ApplyItemReleaseLocations(tx, release, detail, at); err != nil {
				tx.Rollback()
				return release, fiber.StatusUnprocessableEntity, err
			}
		}
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

	// Restore whatever stock this release ever deducted before the cascade delete below
	// removes the detail rows that identify which bins it came from.
	if release.ItemReleaseDetails != nil {
		remarks := fmt.Sprintf("Item Release #%d (deleted)", release.DocNo)
		for _, detail := range *release.ItemReleaseDetails {
			if err := s.RestoreItemReleaseLocations(tx, detail.ID, remarks, at); err != nil {
				tx.Rollback()
				return release, fiber.StatusInternalServerError, err
			}
		}
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

// ApplyItemReleaseLocations persists detail.Locations (the per-bin breakdown the
// dispatching app's PickActivity picker already computes) and actually deducts that
// stock via DeductStockWithTx - the step that was missing entirely before this change.
// Called only from Create/UpdateItemReleaseService, and only for a line whose
// released_qty is new or has just changed (see those callers' own comments).
//
// Requires the bin allocation to sum exactly to ReleasedQty when ReleasedQty > 0,
// rather than silently deducting a partial amount or skipping deduction for a line
// that bypassed the picker (e.g. old client, or a direct grid edit) - see the
// dispatching-side fix that removed released_qty from the warehouse-editable column
// list, closing off the one path that could reach this with an empty allocation.
func (s *ItemReleaseService) ApplyItemReleaseLocations(tx *gorm.DB, release *models.ItemRelease, detail *models.ItemReleaseDetails, at models.At) error {
	if detail.ReleasedQty == 0 {
		return nil
	}

	var allocated uint
	for _, loc := range detail.Locations {
		if loc.SelectedQty > 0 {
			allocated += uint(loc.SelectedQty)
		}
	}
	if allocated != detail.ReleasedQty {
		return fmt.Errorf(
			"item %s: released qty %d does not match the bin allocation total %d - re-pick locations",
			detail.ItemCode, detail.ReleasedQty, allocated,
		)
	}

	remarks := fmt.Sprintf("Item Release #%d", release.DocNo)

	for i := range detail.Locations {
		loc := &detail.Locations[i]
		if loc.SelectedQty <= 0 {
			continue
		}
		loc.ItemReleaseDetailsID = detail.ID

		if err := services.DbInsert(tx, loc); err != nil {
			return fmt.Errorf("failed creating item release location for detail %d: %w", detail.ID, err)
		}

		atdata := models.ItemReleaseLocationsAt{
			RefId:                       loc.ID,
			ItemReleaseLocationsContent: loc.ItemReleaseLocationsContent,
			At:                          at,
		}
		if err := services.DbInsert(tx, &atdata); err != nil {
			return fmt.Errorf("failed creating item release location audit for detail %d: %w", detail.ID, err)
		}

		qty := loc.SelectedQty
		stockBody := &inventory_models.ItemStocks{
			ID: loc.BinId,
			ItemStocksContent: inventory_models.ItemStocksContent{
				StockQty: &qty,
			},
		}
		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   loc.ID,
			SourceType: "item_release",
			Remarks:    remarks,
		}

		if _, err := s.stockService.DeductStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("item %s: %w", detail.ItemCode, err)
		}
	}

	return nil
}

// RestoreItemReleaseLocations reverses whatever ApplyItemReleaseLocations previously
// deducted for one ItemReleaseDetails line and removes those location rows - used both
// on delete and ahead of a re-apply when an already-saved line's released_qty changes.
// A no-op when the line never had a bin allocation (nothing was ever deducted for it).
func (s *ItemReleaseService) RestoreItemReleaseLocations(tx *gorm.DB, itemReleaseDetailsId uint, remarks string, at models.At) error {
	var existing []models.ItemReleaseLocations
	if err := tx.Where("item_release_details_id = ?", itemReleaseDetailsId).Find(&existing).Error; err != nil {
		return fmt.Errorf("failed fetching item release locations for detail %d: %w", itemReleaseDetailsId, err)
	}

	for i := range existing {
		loc := &existing[i]

		qty := loc.SelectedQty
		stockBody := &inventory_models.ItemStocks{
			ID: loc.BinId,
			ItemStocksContent: inventory_models.ItemStocksContent{
				StockQty: &qty,
			},
		}
		stockAtBody := &inventory_models.ItemStocksAt{
			SourceId:   loc.ID,
			SourceType: "item_release",
			Remarks:    remarks,
		}

		if _, err := s.stockService.RestoreStockWithTx(tx, stockBody, stockAtBody, at); err != nil {
			return fmt.Errorf("failed restoring stock for item release location %d: %w", loc.ID, err)
		}
	}

	if err := services.DbDelete(tx, &models.ItemReleaseLocations{}, map[string]interface{}{"item_release_details_id": itemReleaseDetailsId}); err != nil {
		return fmt.Errorf("failed deleting item release locations for detail %d: %w", itemReleaseDetailsId, err)
	}

	return nil
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
