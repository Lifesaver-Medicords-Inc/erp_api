package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
)

type ChartOfAccountService struct{}

func NewChartOfAccountService() *ChartOfAccountService {
	return &ChartOfAccountService{}
}

func (s *ChartOfAccountService) GetChartOfAccounts(conditions map[string]interface{}) ([]accounting_models.ChartOfAccounts, int, error) {
	var response []accounting_models.ChartOfAccounts

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get chart of account")
	}

	return response, fiber.StatusOK, nil
}

func (s *ChartOfAccountService) CreateChartOfAccounts(body *accounting_models.ChartOfAccounts, at models.At) (*accounting_models.ChartOfAccounts, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating chart of accounts")
		}
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record for the main request
	atdata := accounting_models.ChartOfAccountsAt{
		RefId:                 body.ID,
		ChartOfAccountContent: body.ChartOfAccountContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *ChartOfAccountService) UpdateChartOfAccount(body *accounting_models.ChartOfAccounts, conditions map[string]interface{}, at models.At) (*accounting_models.ChartOfAccounts, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record for the main request
	atdata := accounting_models.ChartOfAccountsAt{
		RefId:                 body.ID,
		ChartOfAccountContent: body.ChartOfAccountContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *ChartOfAccountService) DeleteChartOfAccount(body *accounting_models.ChartOfAccounts, at models.At) (*accounting_models.ChartOfAccounts, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record for the main request
	atdata := accounting_models.ChartOfAccountsAt{
		RefId:                 body.ID,
		ChartOfAccountContent: body.ChartOfAccountContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}
