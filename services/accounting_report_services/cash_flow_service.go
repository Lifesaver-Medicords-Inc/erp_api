package accounting_report_services

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
)

// cashAccountIds are excluded from working capital (they ARE the cash the
// statement is explaining, not a use/source of it). Matches CASH ON HAND /
// CASH ON BANK from the live chart (see BalanceSheetService's own
// CashAndOtherAssets, which sums every ASSET account including these two -
// this service pulls them out specifically instead).
var cashAccountIds = map[uint]bool{70034: true, 70035: true}

type CashFlowService struct {
	TrialBalanceService    *TrialBalanceService
	IncomeStatementService *IncomeStatementService
	FixedAssetService      *setup_services.FixedAssetService
}

func NewCashFlowService() *CashFlowService {
	return &CashFlowService{
		TrialBalanceService:    NewTrialBalanceService(),
		IncomeStatementService: NewIncomeStatementService(),
		FixedAssetService:      setup_services.NewFixedAssetService(),
	}
}

// GetCashFlow — Statement of Cash Flows, indirect method, for
// [periodStart, periodEnd]. See CashFlowResult's own doc comment for the
// two things this deliberately doesn't fake (Inventory's change, and
// itemized Financing activities).
func (s *CashFlowService) GetCashFlow(periodStart, periodEnd string) (*accounting_models.CashFlowResult, int, error) {
	startDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_start must be in YYYY-MM-DD format")
	}
	endDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_end must be in YYYY-MM-DD format")
	}
	dayBeforeStart := startDate.AddDate(0, 0, -1).Format("2006-01-02")

	// Point-in-time snapshots (periodStart="" = cumulative since inception,
	// exactly what a Balance Sheet account needs) - one the instant before
	// the period opens, one at the period's close.
	beginRows, status, err := s.TrialBalanceService.GetTrialBalance("", dayBeforeStart)
	if err != nil {
		return nil, status, err
	}
	endRows, status, err := s.TrialBalanceService.GetTrialBalance("", periodEnd)
	if err != nil {
		return nil, status, err
	}

	beginByAccount := map[uint]accounting_models.TrialBalanceRow{}
	for _, r := range beginRows {
		beginByAccount[uint(r.AccountId)] = r
	}
	endByAccount := map[uint]accounting_models.TrialBalanceRow{}
	for _, r := range endRows {
		endByAccount[uint(r.AccountId)] = r
	}

	result := &accounting_models.CashFlowResult{
		PeriodStart:             periodStart,
		PeriodEnd:               periodEnd,
		InventoryChangeExcluded: true,
		FinancingIsResidual:     true,
	}

	// Cash accounts, tracked separately - not part of working capital.
	for id := range cashAccountIds {
		result.BeginningCash += beginByAccount[id].NetBalance
		result.EndingCash += endByAccount[id].NetBalance
	}
	result.NetChangeInCash = result.EndingCash - result.BeginningCash

	// Working capital: every other ASSET/LIABILITY account that shows up on
	// either snapshot. Sign convention matches BalanceSheetService's own:
	// ASSET is debit-normal (NetBalance as-is); LIABILITY is credit-normal
	// (TotalCredit-TotalDebit).
	seen := map[uint]bool{}
	addLine := func(id uint, code, name, class string) {
		if seen[id] || cashAccountIds[id] {
			return
		}
		seen[id] = true

		beginRow := beginByAccount[id]
		endRow := endByAccount[id]

		var beginBal, endBal, cashEffect float64
		switch class {
		case "ASSET":
			beginBal = beginRow.NetBalance
			endBal = endRow.NetBalance
			cashEffect = -(endBal - beginBal) // asset up = cash used
		case "LIABILITY":
			beginBal = beginRow.TotalCredit - beginRow.TotalDebit
			endBal = endRow.TotalCredit - endRow.TotalDebit
			cashEffect = endBal - beginBal // liability up = cash freed
		default:
			return
		}

		result.WorkingCapitalLines = append(result.WorkingCapitalLines, accounting_models.CashFlowWorkingCapitalLine{
			AccountId: id, Code: code, Name: name, AccountClass: class,
			BeginBalance: beginBal, EndBalance: endBal, CashEffect: cashEffect,
		})
		result.NetChangeInWorkingCapital += cashEffect
	}

	for _, r := range beginRows {
		if r.AccountClass == "ASSET" || r.AccountClass == "LIABILITY" {
			addLine(uint(r.AccountId), r.Code, r.Name, r.AccountClass)
		}
	}
	for _, r := range endRows {
		if r.AccountClass == "ASSET" || r.AccountClass == "LIABILITY" {
			addLine(uint(r.AccountId), r.Code, r.Name, r.AccountClass)
		}
	}

	// Net Income + Depreciation come from the Income Statement engine
	// directly rather than being recomputed here, so this can never drift
	// from what that report itself shows for the same period.
	incomeStatement, status, err := s.IncomeStatementService.GetIncomeStatement(periodStart, periodEnd)
	if err != nil {
		return nil, status, err
	}
	result.NetIncome = incomeStatement.NetIncome
	result.DepreciationExpense = incomeStatement.DepreciationExpense

	result.NetCashFromOperating = result.NetIncome + result.DepreciationExpense + result.NetChangeInWorkingCapital

	ppeStart := startDate.Format("01/02/2006")
	ppeEnd := endDate.Format("01/02/2006")
	additions, err := s.FixedAssetService.GetAdditionsInPeriod(ppeStart, ppeEnd)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting PP&E additions")
	}
	result.PpeAdditions = additions
	result.NetCashFromInvesting = -additions

	// Financing is the plug that makes the three sections reconcile to the
	// real, observed change in cash - see CashFlowResult's own comment for
	// why this isn't itemized.
	result.NetCashFromFinancing = result.NetChangeInCash - result.NetCashFromOperating - result.NetCashFromInvesting

	return result, fiber.StatusOK, nil
}
