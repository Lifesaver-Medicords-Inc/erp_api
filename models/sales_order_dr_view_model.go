package models

type SalesOrderDrView struct {
	OrderId        uint    `json:"order_id"`
	DocumentNo     string  `json:"document_no"`
	ProjectName    string  `json:"project_name"`
	TotalAmount    float64 `json:"total_amount"`
	CompanyName    string  `json:"company_name"`
	SalesExecutive string  `json:"sales_executive"`
}

func (SalesOrderDrView) TableName() string {
	return "vw_get_sales_order_dr"
}
