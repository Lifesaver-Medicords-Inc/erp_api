package accounting_models

// CashFlowWorkingCapitalLine — one non-cash ASSET or LIABILITY account's
// contribution to the operating section, indirect method.
type CashFlowWorkingCapitalLine struct {
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
//   - Financing activities has no independently-classified line items -
//     this system has no accounts that distinguish debt issuance/
//     repayment or capital contributions/dividends from any other
//     liability or equity movement (the Chart of Accounts isn't
//     structured for it). NetCashFromFinancing is instead the residual
//     that reconciles Operating + Investing to the real, observed change
//     in the Cash accounts - mathematically exact as a total, but not
//     itemized. FinancingIsResidual flags this.
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

	NetCashFromFinancing float64 `json:"net_cash_from_financing"`
	FinancingIsResidual  bool    `json:"financing_is_residual"`

	BeginningCash   float64 `json:"beginning_cash"`
	EndingCash      float64 `json:"ending_cash"`
	NetChangeInCash float64 `json:"net_change_in_cash"`
}
