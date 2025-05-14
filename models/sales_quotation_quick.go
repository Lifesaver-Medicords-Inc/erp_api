package models

type SalesQuotationQuickContent struct {
	// Parent ID references to Sales Quotation Model
	IsChild         bool    `json:"is_child"`
	BasedId         uint    `json:"based_id"`
	BomId           uint    `json:"bom_id"`
	ItemId          uint    `json:"item_id"`
	Components      string  `json:"components"`
	Model           string  `json:"model"`
	Qty             uint    `json:"qty"`
	UnitOfMeasure   string  `json:"unit_of_measure"`
	ListPrice       float64 `json:"list_price"`
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
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesQuotationQuickContent
	At
}

func (SalesQuotationQuickAt) TableName() string {
	return "z_tbl_trans_sales_quotation_quick_at"
}
