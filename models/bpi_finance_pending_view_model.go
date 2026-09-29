package models

type BpiFinancePendingView struct {
	CustomerId             uint    `json:"customer_id"`
	FinancePendingBranchId uint    `json:"finance_pending_branch_id"`
	Date                   string  `json:"date"`
	QouteRef               string  `json:"qoute_ref"`
	TotalPrice             float64 `json:"total_price"`
	Stage                  string  `json:"stage"`
	Status                 string  `json:"status"`
}

func (BpiFinancePendingView) TableName() string {
	return "vw_get_bpi_finance_pending"
}
