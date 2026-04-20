package inventory_models

type SalesOrderItemReqDetailsView struct {
	SalesOrderDetailsId int    `json:"sales_order_details_id"`
	SalesOrderId        int    `json:"sales_order_id"`
	ItemId              int    `json:"item_id"`
	ItemDesc            string `json:"item_desc"`
	RequiredQty         int    `json:"required_qty"`
	RequiredUom         string `json:"required_uom"`
	RemainingQty        int    `json:"remaining_qty"`
	RemainingUom        string `json:"remaining_uom"`
}

type SalesOrderItemReqDocView struct {
	SalesOrderId uint   `json:"sales_order_id"`
	SoDocNo      string `json:"so_doc_no"`
}

func (SalesOrderItemReqDocView) TableName() string {
	return "vw_get_sales_order_item_req_doc"
}
