package accounting_models

import "github.com/pierceperado/smpc/models"

type BpiOverpaymentContent struct {
	BpiId             uint    `json:"bpi_id"`
	OverpaymentAmount float64 `json:"overpayment_amount"`
	ReferenceDate     string  `json:"reference_date"`
	ReferenceDocType  string  `json:"reference_doc_type"`
	ReferenceDocId    uint    `json:"reference_doc_id"`
}

type BpiOverpayment struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiOverpaymentContent
}

func (BpiOverpayment) TableName() string {
	return "tbl_accounting_bpi_overpayment"
}

type BpiOverpaymentAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiOverpaymentContent
	models.At
}

func (BpiOverpaymentAt) TableName() string {
	return "z_tbl_accounting_bpi_overpayment_at"
}
