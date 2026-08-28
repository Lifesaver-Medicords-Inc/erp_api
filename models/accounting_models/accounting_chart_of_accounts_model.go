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
