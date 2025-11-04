package models

type SalesOrderView struct {
	ItemId          uint   `json:"item_id"`
	RefDoc          string `json:"ref_doc"`
	ItemDescription string `json:"item_description"`
	ReqUom          string `json:"req_uom"`
}

func (SalesOrderView) TableName() string {
	return "vw_get_sales_order_ir"
}
