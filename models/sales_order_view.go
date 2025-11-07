package models

type SalesOrderView struct {
	SOId            uint   `json:"so_id"`
	SODId           uint   `json:"sod_id"`
	ItemId          uint   `json:"item_id"`
	RefDoc          string `json:"ref_doc"`
	ItemDescription string `json:"item_description"`
	OrderQty        uint   `json:"order_qty"`
	ReqQty          uint   `json:"req_qty"`
	ReqUom          string `json:"req_uom"`
}

func (SalesOrderView) TableName() string {
	return "vw_get_sales_order_ir"
}
