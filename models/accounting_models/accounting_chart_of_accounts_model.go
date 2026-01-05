package accounting_models

import "github.com/pierceperado/smpc/models"

type ChartOfAccountContent struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	AccountClass string `json:"account_class"`
	ClassId      uint   `json:"class_id"`
	Group        string `json:"group"`
	GroupId      uint   `json:"group_id"`
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
