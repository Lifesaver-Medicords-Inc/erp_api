package accounting_models

// FinancialRatiosResult — the four standard ratio categories, computed over
// live Balance Sheet (AsOf) and Income Statement (PeriodStart/PeriodEnd)
// data.
//
// Liquidity used to be absent entirely: it needs a Current vs Non-current
// split on the Chart of Accounts, which did not exist, and this model's
// earlier comment recorded that guessing would be worse than not computing.
// §12.10 has since specified that split, and it now lives on
// ChartOfAccountContent.LiquidityClass, so the four figures below compute
// from what management classified - never from an inference.
type FinancialRatiosResult struct {
	AsOf        string `json:"as_of"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`

	// Liquidity - §12.10.
	//
	// LiquidityBlocked is now true only when the classification is missing
	// ENTIRELY on one side (no CURRENT/CASH asset, or no CURRENT liability),
	// leaving nothing to divide. A PARTIAL classification is not blocked: the
	// ratios compute from what is classified and UnclassifiedAccounts names
	// what was left out, which §12.10 requires be visible rather than silent.
	LiquidityBlocked bool   `json:"liquidity_blocked"`
	LiquidityNote    string `json:"liquidity_note"`

	CurrentRatio   float64 `json:"current_ratio"`   // current assets / current liabilities
	QuickRatio     float64 `json:"quick_ratio"`     // (current assets - inventory) / current liabilities
	CashRatio      float64 `json:"cash_ratio"`      // cash and equivalents / current liabilities
	WorkingCapital float64 `json:"working_capital"` // current assets - current liabilities

	// The inputs, exposed so the report can show its working and so the
	// figures can be reconciled against the Balance Sheet by hand.
	CurrentAssets      float64 `json:"current_assets"`
	CurrentLiabilities float64 `json:"current_liabilities"`
	CashAndEquivalents float64 `json:"cash_and_equivalents"`

	// QuickRatioInventory is the figure subtracted for the quick ratio. It is
	// whatever the Balance Sheet reports, so the two cannot disagree - which
	// today means the FIFO inventory module rather than the ledger, because
	// INVENTORY has no postings (see accounting_balance_sheet_model.go).
	QuickRatioInventory float64 `json:"quick_ratio_inventory"`

	// UnclassifiedAccounts: every ASSET or LIABILITY account carrying a
	// balance whose LiquidityClass is still "". These are EXCLUDED from the
	// ratios above and MUST be shown by any consumer - an unclassified
	// account silently treated as current would overstate every ratio, which
	// is the specific failure §12.10 names.
	//
	// Accounts with no balance in the period never reach the trial balance at
	// all (its HAVING clause drops them), and are correctly absent here: an
	// account contributing nothing cannot distort a ratio by being excluded.
	UnclassifiedAccounts []TrialBalanceRow `json:"unclassified_accounts"`

	// Solvency - fully computed, no simplifications.
	DebtToEquity float64 `json:"debt_to_equity"` // Total Liabilities / Total Equity
	DebtRatio    float64 `json:"debt_ratio"`     // Total Liabilities / Total Assets

	// Profitability - fully computed, no simplifications.
	GrossProfitMargin float64 `json:"gross_profit_margin"` // Gross Profit / Revenue
	NetProfitMargin   float64 `json:"net_profit_margin"`   // Net Income / Revenue
	ReturnOnAssets    float64 `json:"return_on_assets"`    // Net Income / Total Assets (as of period end)
	ReturnOnEquity    float64 `json:"return_on_equity"`    // Net Income / Total Equity (as of period end)

	// Efficiency - InventoryTurnover and AssetTurnover both use PERIOD-END
	// balances rather than a period average (a real simplification, flagged
	// below) - Inventory has no historical "as of a past date" snapshot at
	// all to average against (same limitation the Cash Flow Statement
	// flags), and averaging Total Assets around a constant, always-current
	// Inventory figure would be more misleading than using ending balances
	// plainly. Receivables Turnover has no such gap - Trade Receivable is a
	// real ledger account, so its beginning-of-period balance is genuinely
	// computable, and it does use a true period average.
	InventoryTurnover                  float64 `json:"inventory_turnover"` // Cost of Sales / Ending Inventory
	InventoryTurnoverUsesEndingBalance bool    `json:"inventory_turnover_uses_ending_balance"`
	ReceivablesTurnover                float64 `json:"receivables_turnover"` // Revenue / Average Trade Receivable
	AssetTurnover                      float64 `json:"asset_turnover"`       // Revenue / Ending Total Assets
	AssetTurnoverUsesEndingBalance     bool    `json:"asset_turnover_uses_ending_balance"`

	// ReceivablesTurnoverBlocked - Receivables Turnover is the ONLY figure on
	// this report that needs one specific named account rather than an account
	// class, so it is the only one that can be broken by management editing
	// the Chart of Accounts. When the Trade Receivable code is absent the
	// ratio is reported as not computed; it must never fall through to 0.00,
	// which reads as a real (and disastrous) measurement rather than as a
	// missing account. See financial_ratios_service.go's tradeReceivableCode.
	ReceivablesTurnoverBlocked bool   `json:"receivables_turnover_blocked"`
	ReceivablesTurnoverNote    string `json:"receivables_turnover_note"`
}
