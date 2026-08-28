package setup_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

// AssetCategoryService — CRUD for the PP&E register's category level
// (§4.2.3-style Setup list, but for fixed assets - see
// accounting_asset_category_model.go's own comment for why this exists at
// all). Same shape as ChartClassService.
type AssetCategoryService struct{}

func NewAssetCategoryService() *AssetCategoryService {
	return &AssetCategoryService{}
}

func (s *AssetCategoryService) GetAssetCategories(conditions map[string]interface{}, search string, id int) (interface{}, int, utils.PaginationMeta, error) {
	var categories []accounting_models.AssetCategory

	searchColumns := []string{"code", "name"}

	hasNext, pageSize, err := services.DbSearch(&categories, nil, search, searchColumns, nil, id, "id")
	if err != nil {
		return categories, fiber.StatusInternalServerError, utils.PaginationMeta{}, errors.New("failed getting asset categories")
	}

	pagination := utils.PaginationMeta{HasNext: hasNext, PageSize: pageSize}

	return categories, fiber.StatusOK, pagination, nil
}

func (s *AssetCategoryService) GetAssetCategory(id int) (accounting_models.AssetCategory, int, error) {
	var category accounting_models.AssetCategory

	if err := services.DbGet(&category, map[string]interface{}{"id": id}); err != nil {
		return category, fiber.StatusInternalServerError, errors.New("failed getting asset category")
	}

	return category, fiber.StatusOK, nil
}

func (s *AssetCategoryService) CreateAssetCategory(body *accounting_models.AssetCategory, at models.At) (*accounting_models.AssetCategory, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating asset category")
	}

	atdata := accounting_models.AssetCategoryAt{RefId: body.ID, AssetCategoryContent: body.AssetCategoryContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed creating asset category at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *AssetCategoryService) UpdateAssetCategory(body *accounting_models.AssetCategory, at models.At) (*accounting_models.AssetCategory, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating asset category")
	}

	atdata := accounting_models.AssetCategoryAt{RefId: body.ID, AssetCategoryContent: body.AssetCategoryContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed updating asset category at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *AssetCategoryService) DeleteAssetCategory(body *accounting_models.AssetCategory, at models.At) (*accounting_models.AssetCategory, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting asset category")
	}

	atdata := accounting_models.AssetCategoryAt{RefId: body.ID, AssetCategoryContent: body.AssetCategoryContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("failed creating asset category at: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}
