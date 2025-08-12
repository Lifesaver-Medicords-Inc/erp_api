package accounting_models

import "github.com/pierceperado/smpc/models"

type ChartOfAccount struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	AccountCode string `gorm:"unique" json:"account_code"`
	AccountName string `json:"account_name"`
	ShortName   string `json:"short_name"`
	AccountType string `json:"account_type"`
}

func (ChartOfAccount) TableName() string {
	return "tbl_setup_chart_of_account"
}

type ChartOfAccountAt struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	RefId       uint   `json:"ref_id"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	ShortName   string `json:"short_name"`
	AccountType string `json:"account_type"`
	models.At
}

func (ChartOfAccountAt) TableName() string {
	return "z_tbl_setup_chart_of_account_at"
}
