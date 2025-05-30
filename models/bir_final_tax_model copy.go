package models

type FinalTax struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
	Rate string `json:"rate"`
}

func (FinalTax) TableName() string {
	return "tbl_setup_final_tax"
}

type FinalTaxAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Rate  string `json:"rate"`
	At
}

func (FinalTaxAt) TableName() string {
	return "z_tbl_setup_final_tax_at"
}
