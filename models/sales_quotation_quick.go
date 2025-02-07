package models

type SalesQuotationQuickContent struct {
	// Parent ID references to Sales Quotation Model
	BasedId         uint    `json:"based_id"`
	ItemId          uint    `json:"item_id"`
	ItemNameId      uint    `json:"item_name_id"`
	ItemClassId     uint    `json:"item_class_id"`
	Qty             uint    `json:"qty"`
	UnitCode        uint    `json:"unit_code"`
	UnitPrice       float64 `json:"unit_price"`
	PercentDiscount float64 `json:"percent_discount"`
	NetDiscount     float64 `json:"net_discount"`
	NetTotal        float64 `json:"net_total"`
	LineTotal       float64 `json:"line_total"`
}

type SalesQuotationQuick struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesQuotationQuickContent
}

func (SalesQuotationQuick) TableName() string {
	return "tbl_trans_sales_quotation_quick"
}

type SalesQuotationQuickAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	SalesQuotationQuickContent
	At
}

func (SalesQuotationQuickAt) TableName() string {
	return "z_tbl_trans_sales_quotation_quick_at"
}
