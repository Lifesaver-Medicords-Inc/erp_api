package accounting_models

import "github.com/pierceperado/smpc/models"

type TaxContent struct {
	TaxDesc            string `json:"tax_desc"`
	InputTaxCreditable *bool  `json:"input_tax_creditable"`
	CoaSalesId         uint   `json:"coa_sales_id"`
	CoaPurchaseId      uint   `json:"coa_purchase_id"`
	Remarks            string `json:"remarks"`
}

type Tax struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	TaxContent
}

func (Tax) TableName() string {
	return "tbl_setup_tax"
}

type TaxAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	TaxContent
	models.At
}

func (TaxAt) TableName() string {
	return "z_tbl_setup_tax_at"
}
