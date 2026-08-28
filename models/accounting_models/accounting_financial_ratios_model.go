package accounting_models

// FinancialRatiosResult — the four standard ratio categories, computed over
// live Balance Sheet (AsOf) and Income Statement (PeriodStart/PeriodEnd)
// data. Liquidity is deliberately absent: Current Ratio and Quick Ratio
// both need a Current vs Non-current split on the Chart of Accounts, which
// doesn't exist (same gap the Balance Sheet's own model already documents -
// "a forced guess at which accounts are current would be worse than not
// attempting it"). LiquidityBlocked is always true today; nothing computes
// a fake number in its place.
type FinancialRatiosResult struct {
	AsOf        string `json:"as_of"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`

	LiquidityBlocked bool   `json:"liquidity_blocked"`
	LiquidityNote    string `json:"liquidity_note"`

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
}
