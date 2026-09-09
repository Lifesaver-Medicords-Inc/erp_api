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

	// DbUpdate hands GORM a STRUCT, and GORM skips zero-valued fields in a
	// struct update - so "" never reaches the database and a classification
	// can be set but never UNSET. Both columns below are legitimately
	// clearable, so they are rewritten here from a map, which writes every
	// key regardless of whether the value is zero.
	//
	//   - liquidity_class: unclassified is a REAL state under §12.10, not a
	//     missing value. An account classified CURRENT by mistake has to be
	//     returnable to unclassified, or the liquidity ratios keep counting
	//     it forever.
	//   - cash_flow_category: same shape, same problem - an account wrongly
	//     tagged FINANCING could not be untagged, and the Cash Flow
	//     Statement would keep pulling it into its Financing section.
	//
	// Deliberately scoped to this one service rather than fixed inside
	// DbUpdate: teaching that to write zero values would change every setup
	// update in the system, and would start overwriting columns a client
	// never sent.
	clearable := map[string]interface{}{
		"liquidity_class":    body.LiquidityClass,
		"cash_flow_category": body.CashFlowCategory,
	}

	if err := tx.Model(&accounting_models.ChartOfAccounts{}).
		Where("id = ?", body.ID).
		UpdateColumns(clearable).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating chart of account classification")
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
