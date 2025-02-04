package models

type BpiCustomerView struct {
	BpiId        uint   `json:"bpi_id"`
	BranchName   string `json:"branch_name"`
	CustomerCode string `json:"customer_code"`
}

func (BpiCustomerView) TableName() string {
	return "GetBpiCustomer"
}
