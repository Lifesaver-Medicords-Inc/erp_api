package models

type PurchasingCanvassSheetContent struct {
	SupplierId        int     `json:"supplier_id"`
	SupplierName      string  `json:"supplier_name"`
	ContactNo         string  `json:"contact_no"`
	OrderSize         int     `json:"order_size"`
	SupplierStock     int     `json:"supplier_stock"`
	PreviousListPrice float64 `json:"previous_list_price"`
	CurrentListPrice  float64 `json:"current_list_price"`
	NewListPrice      float64 `json:"new_list_price"`
	Discount          string  `json:"discount"`
	NetPrice          float64 `json:"net_price"`
	PriceTrend        string  `json:"price_trend"`
	PriceValidity     int     `json:"price_validity"`
	PaymentTerms      int     `json:"payment_terms"`
	LeadTime          string  `json:"lead_time"`
	ItemId            int     `json:"item_id"`
	StartDate         string  `json:"start_date"`
}

type PurchasingCanvassSheet struct {
	ID uint `gorm:"primarykey" json:"id"`
	PurchasingCanvassSheetContent
}

func (PurchasingCanvassSheet) TableName() string {
	return "tbl_purchasing_canvass_sheet_so"
}

type PurchasingCanvassSheetAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchasingCanvassSheetContent
	At
}

func (PurchasingCanvassSheetAt) TableName() string {
	return "z_tbl_purchasing_canvass_sheet_so_at"
}

type PurchasingCanvassSheetSOView struct {
	PurchasingCanvassSheet
	Status string `json:"status"`
}

func (PurchasingCanvassSheetSOView) TableName() string {
	return "vw_get_purchasing_canvass_sheet_so"
}
