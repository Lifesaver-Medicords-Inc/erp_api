package dispatching_models

// A single cost line (Labor, Vehicle, Fuel, Toll Gate, Insurance, Penalty,
// Others, or a custom type added via "+ADD COST") under a logistics route.
type LogisticsRouteCostContent struct {
	RouteId     uint    `gorm:"not null;index" json:"route_id"`
	CostType    string  `json:"cost_type"`
	Description string  `json:"description"`
	Multiplier  float64 `gorm:"default:0" json:"multiplier"`
	Amount      float64 `gorm:"default:0" json:"amount"`
	ReceiptPath string  `json:"receipt_path"`
}

type LogisticsRouteCost struct {
	ID uint `gorm:"primaryKey" json:"id"`
	LogisticsRouteCostContent
}

func (LogisticsRouteCost) TableName() string {
	return "tbl_dispatching_logistics_route_cost"
}
