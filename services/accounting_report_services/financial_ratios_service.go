package accounting_report_services

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models/accounting_models"
)

// tradeReceivableCoaId - see vw_get_coa_setup / the live chart: id 70032,
// code 100002, "TRADE RECEIVABLE" (confirmed live while fixing Sales
// Invoice's own posting - same account it now correctly debits).
const tradeReceivableCoaId uint = 70032

type FinancialRatiosService struct {
	BalanceSheetService    *BalanceSheetService
	IncomeStatementService *IncomeStatementService
	TrialBalanceService    *TrialBalanceService
}

func NewFinancialRatiosService() *FinancialRatiosService {
	return &FinancialRatiosService{
		BalanceSheetService:    NewBalanceSheetService(),
		IncomeStatementService: NewIncomeStatementService(),
		TrialBalanceService:    NewTrialBalanceService(),
	}
}

func safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// GetFinancialRatios — Solvency, Profitability, and Efficiency, computed
// over the Balance Sheet as of periodEnd and the Income Statement for
// [periodStart, periodEnd]. Liquidity is always blocked - see
// FinancialRatiosResult's own doc comment.
func (s *FinancialRatiosService) GetFinancialRatios(periodStart, periodEnd string) (*accounting_models.FinancialRatiosResult, int, error) {
	startDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", periodEnd); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_end must be in YYYY-MM-DD format")
	}

	balanceSheet, status, err := s.BalanceSheetService.GetBalanceSheet(periodEnd)
	if err != nil {
		return nil, status, err
	}

	incomeStatement, status, err := s.IncomeStatementService.GetIncomeStatement(periodStart, periodEnd)
	if err != nil {
		return nil, status, err
	}

	result := &accounting_models.FinancialRatiosResult{
		AsOf:        periodEnd,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,

		LiquidityBlocked: true,
		LiquidityNote: "Current Ratio and Quick Ratio both need a Current/Non-current split on the " +
			"Chart of Accounts, which doesn't exist - guessing which accounts are current would be " +
			"worse than not computing this at all (same reasoning the Balance Sheet's own model documents).",

		DebtToEquity: safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalEquity),
		DebtRatio:    safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalAssets),

		GrossProfitMargin: safeDiv(incomeStatement.GrossProfit, incomeStatement.Revenue),
		NetProfitMargin:   safeDiv(incomeStatement.NetIncome, incomeStatement.Revenue),
		ReturnOnAssets:    safeDiv(incomeStatement.NetIncome, balanceSheet.TotalAssets),
		ReturnOnEquity:    safeDiv(incomeStatement.NetIncome, balanceSheet.TotalEquity),

		InventoryTurnover:                  safeDiv(incomeStatement.CostOfSales, balanceSheet.Inventory),
		InventoryTurnoverUsesEndingBalance: true,
		AssetTurnover:                      safeDiv(incomeStatement.Revenue, balanceSheet.TotalAssets),
		AssetTurnoverUsesEndingBalance:     true,
	}

	// Receivables Turnover: Trade Receivable is a real ledger account, so
	// unlike Inventory/Total Assets above, a true period average is
	// genuinely computable - fetch its balance the instant before the
	// period opens as well as at period end.
	dayBeforeStart := startDate.AddDate(0, 0, -1).Format("2006-01-02")
	beginRows, status, err := s.TrialBalanceService.GetTrialBalance("", dayBeforeStart)
	if err != nil {
		return nil, status, err
	}
	var beginAR float64
	for _, r := range beginRows {
		if r.AccountId == tradeReceivableCoaId {
			beginAR = r.NetBalance
			break
		}
	}
	var endAR float64
	for _, r := range balanceSheet.AssetAccounts {
		if r.AccountId == tradeReceivableCoaId {
			endAR = r.NetBalance
			break
		}
	}
	averageAR := (beginAR + endAR) / 2
	result.ReceivablesTurnover = safeDiv(incomeStatement.Revenue, averageAR)

	return result, fiber.StatusOK, nil
}
