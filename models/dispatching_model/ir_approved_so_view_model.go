package dispatching_models

type SalesOrderWithApprovedIRView struct {
	SalesOrderID   uint   `json:"sales_order_id"`
	CustomerID     uint   `json:"customer_id"`
	CustomerName   string `json:"customer_name"`
	CustomerCode   string `json:"customer_code"`
	Address        string `json:"address"`
	ShipTypeID     uint   `json:"ship_type_id"`
	TinNo          string `json:"tin_no"`
	SalesOrderNo   string `json:"sales_order_no"`
	ItemReleaseID  uint   `json:"item_release_id"`
	ItemReleaseNo  string `json:"item_release_no"`
	DeliverTo      string `json:"deliver_to"`
	SalesExecutive string `json:"sales_executive"`
}

func (SalesOrderWithApprovedIRView) TableName() string {
	return "vw_get_sales_order_with_approved_ir"
}
