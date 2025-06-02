package models

type ExpandedWithholdingTax struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
	Rate string `json:"rate"`
}

func (ExpandedWithholdingTax) TableName() string {
	return "tbl_setup_ewt"
}

type ExpandedWithholdingTaxAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Rate  string `json:"rate"`
	At
}

func (ExpandedWithholdingTaxAt) TableName() string {
	return "z_tbl_setup_ewt_at"
}
