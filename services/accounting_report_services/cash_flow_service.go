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
	}

	// Cash accounts, tracked separately - not part of working capital.
	for id := range cashAccountIds {
		result.BeginningCash += beginByAccount[id].NetBalance
		result.EndingCash += endByAccount[id].NetBalance
	}
	result.NetChangeInCash = result.EndingCash - result.BeginningCash

	// Sign convention matches BalanceSheetService's own: ASSET/EQUITY are
	// debit/credit per their own normal balance below; LIABILITY is
	// credit-normal (TotalCredit-TotalDebit), same as ASSET is debit-normal
	// (NetBalance as-is).
	balanceFor := func(row accounting_models.TrialBalanceRow, class string) float64 {
		if class == "ASSET" {
			return row.NetBalance
		}
		return row.TotalCredit - row.TotalDebit // LIABILITY, EQUITY
	}

	seen := map[uint]bool{}

	// Working capital: non-cash ASSET accounts, and LIABILITY accounts NOT
	// tagged FINANCING (blank/OPERATING is the default treatment - most
	// payables/accruals are operating by nature).
	addWorkingCapitalLine := func(id uint, code, name, class string) {
		if seen[id] || cashAccountIds[id] {
			return
		}
		seen[id] = true

		beginBal := balanceFor(beginByAccount[id], class)
		endBal := balanceFor(endByAccount[id], class)

		var cashEffect float64
		if class == "ASSET" {
			cashEffect = -(endBal - beginBal) // asset up = cash used
		} else {
			cashEffect = endBal - beginBal // liability up = cash freed
		}

		result.WorkingCapitalLines = append(result.WorkingCapitalLines, accounting_models.CashFlowWorkingCapitalLine{
			AccountId: id, Code: code, Name: name, AccountClass: class,
			BeginBalance: beginBal, EndBalance: endBal, CashEffect: cashEffect,
		})
		result.NetChangeInWorkingCapital += cashEffect
	}

	// Financing: LIABILITY or EQUITY accounts explicitly tagged FINANCING.
	// EQUITY is otherwise excluded entirely from working capital and from
	// financing - an untagged equity account's movement is either driven by
	// Net Income (already counted in Operating via NetIncome itself, so
	// including it here again would double-count it) or by an appropriation
	// this system has no workflow for yet (Statement of Changes in Equity).
	financingSeen := map[uint]bool{}
	addFinancingLine := func(id uint, code, name, class string) {
		if financingSeen[id] {
			return
		}
		financingSeen[id] = true

		beginBal := balanceFor(beginByAccount[id], class)
		endBal := balanceFor(endByAccount[id], class)
		cashEffect := endBal - beginBal // up = cash in (debt/equity raised), down = cash out (repaid/dividends)

		result.FinancingLines = append(result.FinancingLines, accounting_models.CashFlowFinancingLine{
			AccountId: id, Code: code, Name: name, AccountClass: class,
			BeginBalance: beginBal, EndBalance: endBal, CashEffect: cashEffect,
		})
		result.NetCashFromFinancingItemized += cashEffect
	}

	classify := func(rows []accounting_models.TrialBalanceRow) {
		for _, r := range rows {
			id := uint(r.AccountId)
			if cashAccountIds[id] {
				continue
			}
			isFinancing := r.CashFlowCategory == "FINANCING"

			switch r.AccountClass {
			case "ASSET":
				addWorkingCapitalLine(id, r.Code, r.Name, r.AccountClass)
			case "LIABILITY":
				if isFinancing {
					addFinancingLine(id, r.Code, r.Name, r.AccountClass)
				} else {
					addWorkingCapitalLine(id, r.Code, r.Name, r.AccountClass)
				}
			case "EQUITY":
				if isFinancing {
					addFinancingLine(id, r.Code, r.Name, r.AccountClass)
				}
				// untagged EQUITY: deliberately excluded, see addFinancingLine's
				// own doc comment.
			}
		}
	}
	classify(beginRows)
	classify(endRows)

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

	// Financing = itemized (accounts explicitly tagged FINANCING) + a
	// residual that always makes the four sections reconcile exactly to
	// the real, observed change in cash - see CashFlowResult's own comment
	// for why the residual exists and how to shrink it (classify more
	// accounts via Chart of Accounts Setup's Cash Flow Category field).
	result.FinancingResidual = result.NetChangeInCash - result.NetCashFromOperating -
		result.NetCashFromInvesting - result.NetCashFromFinancingItemized
	result.NetCashFromFinancing = result.NetCashFromFinancingItemized + result.FinancingResidual

	return result, fiber.StatusOK, nil
}
