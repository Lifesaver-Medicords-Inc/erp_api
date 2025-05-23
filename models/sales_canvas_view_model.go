package models

type SalesCanvasSheetViewContent struct {
	SupplierBasedId uint    `json:"supplier_based_id"`
	ItemBasedId     uint    `json:"item_based_id"`
	NetPrice        float64 `json:"net_price"`
	Discount        string  `json:"discount"`
	UnitPrice       float64 `json:"unit_price"`
	Validity        string  `json:"validity"`
	LeadTime        uint    `json:"lead_time"`
	RemainingDays   uint
}

type SalesCanvasSheetView struct {
	ID uint `json:"id"`
	SalesCanvasSheetViewContent
}

func (SalesCanvasSheetView) TableName() string {
	return "vw_ComputedCanvasView"
}
