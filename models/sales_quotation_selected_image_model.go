package models

type SalesQuotationSelectedImageContent struct {
	QuotationQuickId uint  `json:"quotation_quick_id"`
	ImageId          uint  `json:"image_id"`
	IsSelected       bool `json:"is_selected"`
}

type SalesQuotationSelectedImage struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesQuotationSelectedImageContent
}

func (SalesQuotationSelectedImage) TableName() string {
	return "tbl_trans_sales_quotation_selected_image"
}

type SalesQuotationSelectedImageAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesQuotationSelectedImageContent
	At
}

func (SalesQuotationSelectedImageAt) TableName() string {
	return "z_tbl_trans_sales_quotation_selected_image_at"
}
