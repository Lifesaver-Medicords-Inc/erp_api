package models

type PaymentTermsContent struct {
	Name string `json:"name"`
}

type PaymentTerms struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	PaymentTermsContent
}

func (PaymentTerms) TableName() string {
	return "tbl_setup_payment_terms"
}

type PaymentTermsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PaymentTermsContent
	At
}

func (PaymentTermsAt) TableName() string {
	return "z_tbl_setup_payment_term_at"
}
