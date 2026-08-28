package accounting_report_services

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/item_stock_services"
)

type IncomeStatementService struct {
	TrialBalanceService *TrialBalanceService
	ItemStockService    *item_stock_services.ItemStockService
}

func NewIncomeStatementService() *IncomeStatementService {
	return &IncomeStatementService{
		TrialBalanceService: NewTrialBalanceService(),
		ItemStockService:    item_stock_services.NewItemStockService(),
	}
}

// GetIncomeStatement is increment 2 of the Admin "Reports" build (approved plan).
// Revenue and Operating Expenses come from the Trial Balance engine restricted to
// [periodStart, periodEnd] (Income Statement accounts are period-only, never
// cumulative - see GetTrialBalance's own doc comment on why periodStart matters).
// Cost of Sales comes from Inventory's actual FIFO consumption records instead of
// the periodic Beginning+Purchases-Ending formula the reference financial
// statements use - see ItemStockService.GetCostOfSales's doc comment for why that
// substitution was made rather than trying to reconstruct a beginning-inventory
// snapshot this system has no mechanism to produce.
//
// Sign convention: REVENUE accounts are credit-normal (TrialBalanceRow.NetBalance
// is debit-positive, so a healthy revenue balance comes back negative) - flipped
// here to TotalCredit-TotalDebit so Revenue reads positive. EXPENSE accounts are
// already debit-normal, so NetBalance is used as-is.
//
// Deliberately NOT grouping into the reference statement's line items (Note 10's
// Vatable/Zero-rated Sales split, Note 12's ~20 expense categories) - that needs
// the report-line-grouping mechanism (gap #6 in the plan), still an open decision.
// RevenueAccounts/ExpenseAccounts carry the real per-account detail in the
// meantime, just not grouped into those categories yet.
func (s *IncomeStatementService) GetIncomeStatement(periodStart, periodEnd string) (*accounting_models.IncomeStatementResult, int, error) {
	if _, err := time.Parse("2006-01-02", periodStart); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", periodEnd); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_end must be in YYYY-MM-DD format")
	}

	rows, status, err := s.TrialBalanceService.GetTrialBalance(periodStart, periodEnd)
	if err != nil {
		return nil, status, err
	}

	result := &accounting_models.IncomeStatementResult{
		PeriodStart:            periodStart,
		PeriodEnd:              periodEnd,
		CostOfSalesIsPerpetual: true,
	}

	for _, row := range rows {
		switch row.AccountClass {
		case "REVENUE":
			result.Revenue += row.TotalCredit - row.TotalDebit
			result.RevenueAccounts = append(result.RevenueAccounts, row)
		case "EXPENSE":
			result.OperatingExpenses += row.NetBalance
			result.ExpenseAccounts = append(result.ExpenseAccounts, row)
		}
	}

	costOfSales, err := s.ItemStockService.GetCostOfSales(periodStart, periodEnd)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting cost of sales")
	}
	result.CostOfSales = costOfSales

	result.GrossProfit = result.Revenue - result.CostOfSales
	result.NetIncome = result.GrossProfit - result.OperatingExpenses

	return result, fiber.StatusOK, nil
}
