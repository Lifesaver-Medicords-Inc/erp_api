package models

type SalesCanvasSheetContent struct {
	NetPrice  float64 `json:"net_price"`
	Discount  string  `json:"discount"`
	UnitPrice float64 `json:"unit_price"`
	Validity  uint    `json:"validity"`
	LeadTime  uint    `json:"lead_time"`
}

type SalesCanvasSheet struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesCanvasSheetContent
}

func (SalesCanvasSheet) TableName() string {
	return "tbl_sales_canvas_sheet"
}

// MUST DO AT SOON
