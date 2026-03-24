package setup_services

import (
	// "errors"

	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

type ChartClassService struct{}

func NewChartClassService() *ChartClassService {
	return &ChartClassService{}
}

func (s *ChartClassService) GetChartClasses(conditions map[string]interface{}, search string, id int) (interface{}, int, utils.PaginationMeta, error) {
	var classes []accounting_models.ChartClass

	searchColumns := []string{
		"code",
		"name",
		"type",
	}

	hasNext, pageSize, err := services.DbSearch(&classes, nil, search, searchColumns, nil, id, "id")
	if err != nil {
		return classes, fiber.StatusInternalServerError, utils.PaginationMeta{}, errors.New("failed getting chart classes")
	}

	pagination := utils.PaginationMeta{
		HasNext:  hasNext,
		PageSize: pageSize,
	}

	return classes, fiber.StatusOK, pagination, nil
}

func (s *ChartClassService) GetChartClass(id int) (accounting_models.ChartClass, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var class accounting_models.ChartClass

	if err := services.DbGet(&class, conditions); err != nil {
		return class, fiber.StatusInternalServerError, errors.New("failed getting class")
	}

	return class, fiber.StatusOK, nil
}

func (s *ChartClassService) CreateChartClass(body *accounting_models.ChartClass, at models.At) (*accounting_models.ChartClass, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	// Insert main Chart Class
	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating chart class")
	}

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating chart class at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *ChartClassService) UpdateChartClass(body *accounting_models.ChartClass, conditions map[string]interface{}, at models.At) (*accounting_models.ChartClass, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating class")
	}

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *ChartClassService) DeleteChartClass(body *accounting_models.ChartClass, at models.At) (*accounting_models.ChartClass, int, error) {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting class")
	}

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}
