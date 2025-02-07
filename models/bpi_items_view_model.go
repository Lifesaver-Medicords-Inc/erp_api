package models

type BpiItemsViewContent struct {
	BpiItemBasedId uint    `json:"bpi_item_based_id"`
	PaymentTermsId string  `json:"payment_terms_id"`
	TaxCode        string  `json:"tax_code"`
	ItemTaxCode    string  `json:"item_tax_code"`
	Price          float64 `json:"price"`
	Notes          string  `json:"notes"`
	ItemId         uint    `json:"item_id"`
	ItemCode       string  `json:"item_code"`
	ShortDesc      string  `json:"short_desc"`
	StatusTangible string  `json:"status_tangible"`
	StatusTrade    string  `json:"status_trade"`
}

type BpiItemsView struct {
	BpiItemID uint `gorm:"primarykey" json:"bpi_item_id"`
	BpiItemsViewContent
}

func (BpiItemsView) TableName() string {
	return "vw_bpi_items"
}
