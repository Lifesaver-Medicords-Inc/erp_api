package accounting_report_services

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/accounting_models"
)

type TrialBalanceService struct{}

func NewTrialBalanceService() *TrialBalanceService {
	return &TrialBalanceService{}
}

// GetTrialBalance sums every posted journal entry detail line, per account, up to
// and including asOfDate (expected as "YYYY-MM-DD", the caller's responsibility -
// see the handler), optionally not-before periodStart. This is increment 1 of the
// Admin "Reports" build (approved plan) - the aggregation every later statement
// composes on top of, not a report in its own right yet, which is why there's no
// UI for it - verified directly against live data instead.
//
// periodStart is what makes this one engine serve both kinds of report: pass ""
// for a cumulative since-inception balance (correct for Balance Sheet accounts -
// Assets/Liabilities/Equity are running balances, not reset each period) or a
// real date for period-only activity (correct for Income Statement accounts -
// Revenue/Expense must be measured for the period alone, never all-time).
//
// tbl_accounting_journal_entry_details.posting_date is a free-typed string, not a
// real date column (mirrors JournalEntry.period_from/period_to, fixed earlier this
// session for the exact same reason) - stored MM/dd/yyyy, matching the layout
// every auto-posting call site parses DocDate with (time.Parse("01/02/2006", ...)
// in sales_invoice_service.go and siblings). TRY_CONVERT(..., 101) reads it as
// that style; a row whose posting_date somehow isn't valid MM/dd/yyyy is silently
// excluded (TRY_CONVERT -> NULL, fails the comparison) rather than crashing the
// whole report over one bad row. The caller's dates are converted separately with
// style 23 (ISO yyyy-mm-dd) so their format never has to match the legacy stored
// one.
//
// Money is cast to DECIMAL(18,2) for the SUM itself: debit/credit are plain
// `float` columns (float64 in Go) throughout this codebase - summing hundreds of
// journal lines in binary float risks a visible off-by-a-centavo drift a
// filed-style financial report can't afford. This casts only at the aggregation
// boundary; it doesn't touch the underlying column type or any existing posting
// code.
func (s *TrialBalanceService) GetTrialBalance(periodStart, asOfDate string) ([]accounting_models.TrialBalanceRow, int, error) {
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("as_of must be in YYYY-MM-DD format")
	}
	if periodStart != "" {
		if _, err := time.Parse("2006-01-02", periodStart); err != nil {
			return nil, fiber.StatusBadRequest, errors.New("period_start must be in YYYY-MM-DD format")
		}
	}

	var response []accounting_models.TrialBalanceRow

	query := `
		SELECT
			coa.id AS account_id,
			coa.code,
			coa.name,
			ISNULL(cc.type, '') AS account_class,
			CAST(ISNULL(SUM(CAST(jed.debit AS DECIMAL(18,2))), 0) AS FLOAT) AS total_debit,
			CAST(ISNULL(SUM(CAST(jed.credit AS DECIMAL(18,2))), 0) AS FLOAT) AS total_credit,
			CAST(ISNULL(SUM(CAST(jed.debit AS DECIMAL(18,2))), 0)
			   - ISNULL(SUM(CAST(jed.credit AS DECIMAL(18,2))), 0) AS FLOAT) AS net_balance,
			ISNULL(coa.cash_flow_category, '') AS cash_flow_category
		FROM tbl_setup_chart_of_accounts coa
		LEFT JOIN tbl_setup_chart_class cc ON coa.class_id = cc.id
		LEFT JOIN tbl_accounting_journal_entry_details jed
			ON jed.posting_ref_id = coa.id
			AND TRY_CONVERT(date, jed.posting_date, 101) <= TRY_CONVERT(date, ?, 23)
			AND (? = '' OR TRY_CONVERT(date, jed.posting_date, 101) >= TRY_CONVERT(date, ?, 23))
		GROUP BY coa.id, coa.code, coa.name, cc.type, coa.cash_flow_category
		HAVING ISNULL(SUM(jed.debit), 0) <> 0 OR ISNULL(SUM(jed.credit), 0) <> 0
		ORDER BY coa.code
	`

	if err := initializers.DB.Raw(query, asOfDate, periodStart, periodStart).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting trial balance")
	}

	return response, fiber.StatusOK, nil
}
