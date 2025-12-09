package models

type ValuationMethodContent struct {
	Name string `json:"name"`
}

type ValuationMethod struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	ValuationMethodContent
}

func (ValuationMethod) TableName() string {
	return "tbl_setup_valuation_method"
}

type ValuationMethodAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ValuationMethodContent
	At
}

func (ValuationMethodAt) TableName() string {
	return "z_tbl_setup_valuation_method_at"
}
