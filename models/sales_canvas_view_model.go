package models

type SalesCanvasSheetViewContent struct {
	StartDate      string `json:"start_date"`
	Validity       uint
	TotalDays      uint
	ExpirationDate string
	RemainingDays  uint
	Status         string
}

type SalesCanvasSheetView struct {
	ID uint `json:"id"`
}

func (SalesCanvasSheetView) TableName() string {
	return "vw_CanvasSheet"
}
