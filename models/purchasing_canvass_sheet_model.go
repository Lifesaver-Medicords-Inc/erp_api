package models

type PurchasingCanvassSheetContent struct {
	SupplierId       int     `json:"supplier_id"`
	SupplierName     string  `json:"supplier_name"`
	ContactNo        string  `json:"contact_no"`
	OrderSize        int     `json:"order_size"`
	SupplierStock    int     `json:"supplier_stock"`
	CurrentListPrice float64 `json:"current_list_price"`
	NewListPrice     float64 `json:"new_list_price"`
	Discount         float64 `json:"discount"`
	NetPrice         float64 `json:"net_price"`
	PriceValidity    string  `json:"price_validity"`
	PaymentTerms     float64 `json:"payment_terms"`
	LeadTime         float64 `json:"lead_time"`
	ItemId           int     `json:"item_id"`
	Date             string  `json:"date"`
}

type PurchasingCanvassSheet struct {
	ID uint `gorm:"primarykey" json:"id"`
	PurchasingCanvassSheetContent
}

func (PurchasingCanvassSheet) TableName() string {
	return "tbl_purchasing_canvass_sheet"
}

type PurchasingCanvassSheetAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchasingCanvassSheetContent
	At
}

func (PurchasingCanvassSheetAt) TableName() string {
	return "z_tbl_purchasing_canvass_sheet_at"
}
