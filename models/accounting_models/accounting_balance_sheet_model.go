package accounting_models

// BalanceSheetResult is increment 3 of the Admin "Reports" build (approved plan).
// Deliberately NOT split into Current/Non-current (gap #1 in the plan - the Chart
// of Accounts has no such sub-classification, only the 5 top-level classes) - a
// forced guess at which accounts are "current" would be worse than not
// attempting it. AssetAccounts/LiabilityAccounts/EquityAccounts carry the real
// per-account detail instead, same reasoning as the Income Statement's
// RevenueAccounts/ExpenseAccounts.
type BalanceSheetResult struct {
	AsOf string `json:"as_of"`

	// CashAndOtherAssets is every ASSET-class ledger account summed (Cash, Trade
	// Receivable, etc.) - it does NOT include Inventory, which this system never
	// posts to the ledger at all (see InventoryIsFromInventoryModule below).
	CashAndOtherAssets float64 `json:"cash_and_other_assets"`
	// Inventory comes from ItemStockService.GetInventoryValue (the Inventory
	// module's own FIFO lot data), not from any Chart-of-Accounts balance - this
	// system has no "Inventory" GL account, and Invoice Receipt / Receiving
	// Report don't post one when stock comes in (confirmed: Invoice Receipt's
	// auto-posting debits ACCRUED EXPENSE PAYABLE and credits INVENTORY GAIN -
	// neither is an asset account. Flagged as its own discrepancy, not fixed
	// here). So there's no double-counting risk combining the two sources, but
	// also no reason to expect the combined Balance Sheet to actually balance
	// against Liabilities+Equity the way GlAssetsEqualsLiabilitiesPlusEquity
	// (GL-only, excluding this field) does.
	Inventory                      float64 `json:"inventory"`
	InventoryIsFromInventoryModule bool    `json:"inventory_is_from_inventory_module"`

	// PropertyAndEquipment is always 0 today - this system has no fixed-asset or
	// depreciation register at all (gap #3 in the plan; FS 2023's Note 7 shows a
	// real ₱12.85M net carrying amount with nothing here to compute it from).
	// PropertyAndEquipmentIsTracked is false so the UI can show "not tracked"
	// instead of a misleading ₱0.00.
	PropertyAndEquipment         float64 `json:"property_and_equipment"`
	PropertyAndEquipmentIsTracked bool   `json:"property_and_equipment_is_tracked"`

	TotalAssets float64 `json:"total_assets"`

	TotalLiabilities          float64 `json:"total_liabilities"`
	TotalEquity               float64 `json:"total_equity"`
	TotalLiabilitiesAndEquity float64 `json:"total_liabilities_and_equity"`

	AssetAccounts     []TrialBalanceRow `json:"asset_accounts"`
	LiabilityAccounts []TrialBalanceRow `json:"liability_accounts"`
	EquityAccounts    []TrialBalanceRow `json:"equity_accounts"`

	// GlAssetsEqualsLiabilitiesPlusEquity checks CashAndOtherAssets alone against
	// TotalLiabilities+TotalEquity - i.e. whether the ledger itself is internally
	// consistent, before Inventory or PP&E (neither of which the ledger knows
	// about) are added in. False is a real red flag about the ledger's postings,
	// not about this report.
	GlAssetsEqualsLiabilitiesPlusEquity bool `json:"gl_assets_equals_liabilities_plus_equity"`
}
