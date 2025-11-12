package models

type SalesOrderViewIR struct {
	SOId            uint   `json:"so_id"`
	SODId           uint   `json:"sod_id"`
	ItemId          uint   `json:"item_id"`
	RefDoc          string `json:"ref_doc"`
	ItemDescription string `json:"item_description"`
	OrderQty        uint   `json:"order_qty"`
	ReqQty          uint   `json:"req_qty"`
	ReqUom          string `json:"req_uom"`
}

func (SalesOrderViewIR) TableName() string {
	return "vw_get_sales_order_ir"
}

type SalesOrderViewPA struct {
	SOId            uint   `json:"so_id"`
	SODId           uint   `json:"sod_id"`
	ItemId          uint   `json:"item_id"`
	RefDoc          string `json:"ref_doc"`
	Customer        string `json:"customer"`
	CustomerCode    string `json:"customer_code"`
	SalesPerson     string `json:"sales_person"`
	ItemDescription string `json:"item_description"`
	LeftQty         uint   `json:"left_qty"`
	PickQty         uint   `json:"_qty"`
	LeftUom         string `json:"left_uom"`
}

func (SalesOrderViewPA) TableName() string {
	return "vw_get_sales_order_pa"
}
