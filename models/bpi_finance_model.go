package models

type BpiFinanceContent struct {
	FinanceBasedId        uint   `json:"finance_based_id"`
	FinanceAccountId      uint   `json:"finance_account_id"`
	FinancePaymentTermsId uint   `json:"finance_payment_terms_id"`
	FinanceBranchId       uint   `json:"finance_branch_id"`
	FinanceTaxCode        string `json:"finance_tax_code"`
	FinanceTax            string `json:"finance_tax"`
}

type BpiFinance struct {
	FinanceID uint `gorm:"primarykey" json:"finance_id"`
	BpiFinanceContent
}

func (BpiFinance) TableName() string {
	return "tbl_bpi_finance"
}

type BpiFinanceAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiFinanceContent
	At
}

func (BpiFinanceAt) TableName() string {
	return "z_tbl_bpi_finance_at"
}
