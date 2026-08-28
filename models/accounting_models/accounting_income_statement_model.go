package accounting_models

// IncomeStatementResult is increment 2 of the Admin "Reports" build (approved
// plan). Shape mirrors FS 2023's actual Income Statement page (Revenue, Cost of
// Sales, Gross Profit, Operating Expenses, Net Income) - see the plan's PDF notes.
//
// RevenueAccounts/ExpenseAccounts are the raw per-account lines behind Revenue/
// OperatingExpenses (straight from the Trial Balance engine, unGrouped) - real
// line-item detail without needing the report-line-grouping mechanism (gap #6,
// still deferred) to reproduce the reference statement's grouped categories
// (Note 10's Vatable/Zero-rated split, Note 12's ~20 expense categories).
type IncomeStatementResult struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`

	Revenue           float64 `json:"revenue"`
	CostOfSales       float64 `json:"cost_of_sales"`
	GrossProfit       float64 `json:"gross_profit"`
	OperatingExpenses float64 `json:"operating_expenses"`
	NetIncome         float64 `json:"net_income"`

	// DepreciationExpense is computed straight-line from tbl_fixed_asset for
	// this exact period (FixedAssetService.GetDepreciationExpense) - not a
	// ledger account, same "compute at read time" reasoning as CostOfSales.
	// Included inside OperatingExpenses, not on top of it.
	DepreciationExpense float64 `json:"depreciation_expense"`

	RevenueAccounts []TrialBalanceRow `json:"revenue_accounts"`
	ExpenseAccounts []TrialBalanceRow `json:"expense_accounts"`

	// CostOfSalesIsPerpetual flags that CostOfSales came from actual FIFO
	// consumption records (ItemStockService.GetCostOfSales), not the periodic
	// Beginning+Purchases-Ending formula the reference financial statements use -
	// see that function's own doc comment for why. Always true today; kept as an
	// explicit field rather than an assumption in case a periodic fallback is
	// ever added alongside it.
	CostOfSalesIsPerpetual bool `json:"cost_of_sales_is_perpetual"`
}
