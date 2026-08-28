package setup_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

// FixedAssetService — CRUD for individual PP&E items, plus the straight-
// line depreciation math the Balance Sheet and Income Statement read from
// (see GetPPEAsOf / GetDepreciationExpense). No depreciation "run" or
// close step exists - both are computed fresh from tbl_fixed_asset on
// every call, the same approach GetCostOfSales already takes against the
// FIFO stock-lot tables rather than a periodic-inventory formula.
type FixedAssetService struct{}

func NewFixedAssetService() *FixedAssetService {
	return &FixedAssetService{}
}

const dateLayout = "01/02/2006"

func (s *FixedAssetService) GetFixedAssets(conditions map[string]interface{}, search string, id int) (interface{}, int, utils.PaginationMeta, error) {
	var assets []accounting_models.FixedAsset

	searchColumns := []string{"code", "name", "category_name", "status"}

	hasNext, pageSize, err := services.DbSearch(&assets, nil, search, searchColumns, nil, id, "id")
	if err != nil {
		return assets, fiber.StatusInternalServerError, utils.PaginationMeta{}, errors.New("failed getting fixed assets")
	}

	pagination := utils.PaginationMeta{HasNext: hasNext, PageSize: pageSize}

	return assets, fiber.StatusOK, pagination, nil
}

func (s *FixedAssetService) GetFixedAsset(id int) (accounting_models.FixedAsset, int, error) {
	var asset accounting_models.FixedAsset

	if err := services.DbGet(&asset, map[string]interface{}{"id": id}); err != nil {
		return asset, fiber.StatusInternalServerError, errors.New("failed getting fixed asset")
	}

	return asset, fiber.StatusOK, nil
}

func (s *FixedAssetService) CreateFixedAsset(body *accounting_models.FixedAsset, at models.At) (*accounting_models.FixedAsset, int, error) {
	if strings.TrimSpace(body.Status) == "" {
		body.Status = "ACTIVE"
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating fixed asset")
	}

	atdata := accounting_models.FixedAssetAt{RefId: body.ID, FixedAssetContent: body.FixedAssetContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed creating fixed asset at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *FixedAssetService) UpdateFixedAsset(body *accounting_models.FixedAsset, at models.At) (*accounting_models.FixedAsset, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating fixed asset")
	}

	atdata := accounting_models.FixedAssetAt{RefId: body.ID, FixedAssetContent: body.FixedAssetContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed updating fixed asset at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *FixedAssetService) DeleteFixedAsset(body *accounting_models.FixedAsset, at models.At) (*accounting_models.FixedAsset, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting fixed asset")
	}

	atdata := accounting_models.FixedAssetAt{RefId: body.ID, FixedAssetContent: body.FixedAssetContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed creating fixed asset at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

// ── Depreciation math ───────────────────────────────────────────────────

// monthsBetween counts whole months from start to end, floored - a partial
// month (end's day-of-month earlier than start's) doesn't count yet. Never
// negative.
func monthsBetween(start, end time.Time) int {
	months := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())
	if end.Day() < start.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

// accumulatedDepreciation is straight-line, capped at (cost - salvage) so
// an asset never depreciates past its salvage value regardless of age.
// UsefulLifeYears <= 0 marks a non-depreciable asset (e.g. Land) - always
// zero, never a divide-by-zero. Returns 0 for a date before acquisition.
func accumulatedDepreciation(asset accounting_models.FixedAsset, asOf time.Time) float64 {
	if asset.UsefulLifeYears <= 0 {
		return 0
	}

	acquired, err := time.Parse(dateLayout, strings.TrimSpace(asset.AcquiredDate))
	if err != nil || asOf.Before(acquired) {
		return 0
	}

	depreciableBase := asset.Cost - asset.SalvageValue
	if depreciableBase <= 0 {
		return 0
	}

	monthlyDep := depreciableBase / (asset.UsefulLifeYears * 12)
	accumulated := monthlyDep * float64(monthsBetween(acquired, asOf))

	if accumulated > depreciableBase {
		accumulated = depreciableBase
	}
	return accumulated
}

// activeAssetsAsOf returns every asset that existed and had not yet been
// disposed as of the given date - a disposed asset drops out of the
// rollup entirely from its disposal date forward, rather than freezing
// its depreciation (a disposal removes it from the books).
func activeAssetsAsOf(asOf time.Time) ([]accounting_models.FixedAsset, error) {
	var assets []accounting_models.FixedAsset
	if err := initializers.DB.Find(&assets).Error; err != nil {
		return nil, err
	}

	var result []accounting_models.FixedAsset
	for _, a := range assets {
		acquired, err := time.Parse(dateLayout, strings.TrimSpace(a.AcquiredDate))
		if err != nil || acquired.After(asOf) {
			continue
		}
		if strings.EqualFold(a.Status, "DISPOSED") && strings.TrimSpace(a.DisposedDate) != "" {
			disposed, err := time.Parse(dateLayout, strings.TrimSpace(a.DisposedDate))
			if err == nil && !disposed.After(asOf) {
				continue
			}
		}
		result = append(result, a)
	}
	return result, nil
}

// GetPPEAsOf — the Balance Sheet's Property & Equipment line, broken down
// by category the same way a real PP&E note is.
func (s *FixedAssetService) GetPPEAsOf(asOfDate string) (accounting_models.PPEResult, error) {
	result := accounting_models.PPEResult{AsOf: asOfDate}

	asOf, err := time.Parse(dateLayout, strings.TrimSpace(asOfDate))
	if err != nil {
		return result, errors.New("invalid as-of date format")
	}

	assets, err := activeAssetsAsOf(asOf)
	if err != nil {
		return result, err
	}

	byCategory := map[uint]*accounting_models.PPECategoryBreakdown{}
	for _, a := range assets {
		accDep := accumulatedDepreciation(a, asOf)
		nbv := a.Cost - accDep

		cat, ok := byCategory[a.CategoryId]
		if !ok {
			cat = &accounting_models.PPECategoryBreakdown{CategoryId: a.CategoryId, CategoryName: a.CategoryName}
			byCategory[a.CategoryId] = cat
		}
		cat.Cost += a.Cost
		cat.AccumulatedDepreciation += accDep
		cat.NetBookValue += nbv

		result.TotalCost += a.Cost
		result.TotalAccumulatedDepreciation += accDep
		result.TotalNetBookValue += nbv
	}

	for _, cat := range byCategory {
		result.Categories = append(result.Categories, *cat)
	}

	return result, nil
}

// GetDepreciationExpense — the Income Statement's depreciation line for one
// period: how much accumulated depreciation grew between the day before
// periodStart and periodEnd. An asset acquired mid-period naturally
// contributes only its own partial-period share, since its "before"
// accumulated depreciation is computed from its own acquisition date.
func (s *FixedAssetService) GetDepreciationExpense(periodStart, periodEnd string) (float64, error) {
	start, err := time.Parse(dateLayout, strings.TrimSpace(periodStart))
	if err != nil {
		return 0, errors.New("invalid period_start format")
	}
	end, err := time.Parse(dateLayout, strings.TrimSpace(periodEnd))
	if err != nil {
		return 0, errors.New("invalid period_end format")
	}

	assets, err := activeAssetsAsOf(end)
	if err != nil {
		return 0, err
	}

	dayBeforeStart := start.AddDate(0, 0, -1)

	var expense float64
	for _, a := range assets {
		expense += accumulatedDepreciation(a, end) - accumulatedDepreciation(a, dayBeforeStart)
	}
	return expense, nil
}
