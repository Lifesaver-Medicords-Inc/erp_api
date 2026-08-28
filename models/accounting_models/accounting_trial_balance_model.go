package accounting_models

// TrialBalanceRow is one Chart-of-Accounts account's aggregated activity as of a
// given date - the foundation every financial report in the Admin "Reports" build
// composes on top of (Balance Sheet, Income Statement, Statement of Cash Flows,
// Statement of Changes in Equity, then the 4 ratio categories - see the approved
// plan for that work).
//
// Deliberately unopinionated about debit/credit-normal presentation: it just sums
// what actually posted. NetBalance = TotalDebit - TotalCredit (debit-positive)
// always; whichever report consumes this decides how to display that for its own
// account class (e.g. a Balance Sheet flips the sign for LIABILITY/EQUITY/REVENUE
// accounts, which are credit-normal).
type TrialBalanceRow struct {
	AccountId uint   `json:"account_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	// AccountClass comes from tbl_setup_chart_class.type via class_id (ASSET /
	// LIABILITY / EQUITY / REVENUE / EXPENSE, singular - matches the 5 values
	// ChartClassPage.cs hardcodes) - NOT the free-text account_class column that
	// also lives directly on the chart-of-accounts row, which isn't a foreign key
	// and can drift from the real classification.
	AccountClass string  `json:"account_class"`
	TotalDebit   float64 `json:"total_debit"`
	TotalCredit  float64 `json:"total_credit"`
	NetBalance   float64 `json:"net_balance"`
	// CashFlowCategory: "" (OPERATING) | "FINANCING" - straight off
	// tbl_setup_chart_of_accounts.cash_flow_category, set via Chart of
	// Accounts Setup. Only the Cash Flow Statement reads this today.
	CashFlowCategory string `json:"cash_flow_category"`
}
