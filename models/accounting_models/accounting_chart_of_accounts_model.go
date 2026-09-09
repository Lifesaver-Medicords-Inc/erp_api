package accounting_models

import "github.com/pierceperado/smpc/models"

type ChartOfAccountContent struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	AccountClass string `json:"account_class"`
	ClassId      uint   `json:"class_id"`
	Group        string `json:"group"`
	GroupId      uint   `json:"group_id"`

	// CashFlowCategory: "" (default - treated as OPERATING) | "FINANCING".
	// Set per account via Chart of Accounts Setup. Only LIABILITY/EQUITY
	// accounts tagged FINANCING get pulled out of the Cash Flow Statement's
	// working-capital section into its itemized Financing section - see
	// cash_flow_service.go. Nothing defaults an account to FINANCING on its
	// own; the user classifies each one, same "GL account chosen by the
	// user" precedent as everywhere else in this system.
	CashFlowCategory string `json:"cash_flow_category"`

	// LiquidityClass: "" (unclassified) | "CURRENT" | "NON-CURRENT" | "CASH".
	// Set per account via Chart of Accounts Setup; §12.10's liquidity ratios
	// read nothing else to decide what is current.
	//
	// §12.10 is explicit that this MUST NOT be inferred from the account's
	// code, class or name: which assets and liabilities are current is a
	// judgement about how this company operates, and it is management's to
	// make. Same rule CashFlowCategory states above, for the same reason.
	//
	// "" is a real state, not a missing value. An unclassified account is
	// EXCLUDED from the ratios and reported as unclassified - never quietly
	// counted as current, which would overstate every ratio. Nothing
	// backfills it.
	//
	// "CASH" means cash and cash equivalents. It is a subset of CURRENT, not
	// an alternative to it: a CASH account counts toward current assets AND
	// toward the cash ratio's numerator. Keeping it on this one field means
	// one decision per account rather than two, and avoids depending on the
	// GL Mapping table (§12.8), which does not exist yet.
	//
	// Named for the classification rather than for the ratios because
	// §12.11.1's Balance Sheet wants the same current/non-current split and
	// should read this field rather than gain a second one.
	LiquidityClass string `json:"liquidity_class"`
}

type ChartOfAccounts struct {
	ID uint `gorm:"primarykey" json:"id"`
	ChartOfAccountContent
}

func (ChartOfAccounts) TableName() string {
	return "tbl_setup_chart_of_accounts"
}

type ChartOfAccountsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ChartOfAccountContent
	models.At
}

func (ChartOfAccountsAt) TableName() string {
	return "z_tbl_setup_chart_of_accounts_at"
}
