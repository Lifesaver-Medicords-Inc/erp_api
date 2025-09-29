package models

type BpiCustomerView struct {
	BpiId        uint   `json:"bpi_id"`
	CustomerCode string `json:"customer_code"`
	BranchName   string `json:"branch_name"`
}

func (BpiCustomerView) TableName() string {
	return "vw_get_bpi_customers"
}
