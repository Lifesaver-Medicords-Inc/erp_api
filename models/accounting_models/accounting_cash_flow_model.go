package accounting_models

// CashFlowWorkingCapitalLine — one non-cash, OPERATING-classified ASSET or
// LIABILITY account's contribution to the operating section, indirect
// method.
type CashFlowWorkingCapitalLine struct {
	AccountId    uint    `json:"account_id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	AccountClass string  `json:"account_class"`
	BeginBalance float64 `json:"begin_balance"`
	EndBalance   float64 `json:"end_balance"`
	CashEffect   float64 `json:"cash_effect"`
}

// CashFlowFinancingLine — one LIABILITY or EQUITY account explicitly
// tagged FINANCING (Chart of Accounts Setup's Cash Flow Category field),
// itemized rather than folded into the residual.
type CashFlowFinancingLine struct {
	AccountId    uint    `json:"account_id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	AccountClass string  `json:"account_class"`
	BeginBalance float64 `json:"begin_balance"`
	EndBalance   float64 `json:"end_balance"`
	CashEffect   float64 `json:"cash_effect"`
}

// CashFlowResult — Statement of Cash Flows, indirect method, for one
// period. Two things are honestly incomplete rather than faked:
//
//   - Inventory's change is NOT in WorkingCapitalLines. Inventory comes
//     from the Inventory module's current FIFO stock value
//     (ItemStockService.GetInventoryValue), which has no "as of a past
//     date" capability - only a current snapshot exists, so there is no
//     beginning-of-period figure to diff against. InventoryChangeExcluded
//     flags this rather than silently treating the change as zero.
//   - Financing is itemized only for accounts the user has explicitly
//     tagged FINANCING via Chart of Accounts Setup's Cash Flow Category
//     field (FinancingLines) - nothing infers this classification on its
//     own. FinancingResidual is whatever's left over after those itemized
//     lines: the amount still needed to reconcile Operating + Investing +
//     itemized Financing to the real, observed change in the Cash
//     accounts. It trends toward zero as more accounts get classified,
//     and is mathematically exact as a total either way - it's the
//     itemization that's incomplete, never the reconciliation.
type CashFlowResult struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`

	NetIncome           float64 `json:"net_income"`
	DepreciationExpense float64 `json:"depreciation_expense"`

	WorkingCapitalLines       []CashFlowWorkingCapitalLine `json:"working_capital_lines"`
	NetChangeInWorkingCapital float64                      `json:"net_change_in_working_capital"`
	InventoryChangeExcluded   bool                         `json:"inventory_change_excluded"`
	NetCashFromOperating      float64                      `json:"net_cash_from_operating"`

	PpeAdditions         float64 `json:"ppe_additions"`
	NetCashFromInvesting float64 `json:"net_cash_from_investing"`

	FinancingLines               []CashFlowFinancingLine `json:"financing_lines"`
	NetCashFromFinancingItemized float64                 `json:"net_cash_from_financing_itemized"`
	FinancingResidual            float64                 `json:"financing_residual"`
	NetCashFromFinancing         float64                 `json:"net_cash_from_financing"` // itemized + residual

	BeginningCash   float64 `json:"beginning_cash"`
	EndingCash      float64 `json:"ending_cash"`
	NetChangeInCash float64 `json:"net_change_in_cash"`
}
