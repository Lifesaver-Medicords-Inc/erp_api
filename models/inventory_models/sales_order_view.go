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

type SalesOrderPickActView struct {
	SalesOrderId int    `json:"sales_order_id"`
	Customer     string `json:"customer"`
	CustomerCode string `json:"customer_code"`
	SalesPerson  string `json:"sales_person"`
}

type SalesOrderPickActDetailsView struct {
	SalesOrderDetailsId int    `json:"sales_order_details_id"`
	SalesOrderId        int    `json:"sales_order_id"`
	ItemId              int    `json:"item_id"`
	ItemCode            string `json:"item_code"`
	ItemDescription     string `json:"item_description"`
	LeftQty             int    `json:"left_qty"`
	LeftUom             string `json:"left_uom"`
}

type SalesOrderPickActDocView struct {
	SalesOrderId uint   `json:"sales_order_id"`
	SoDocNo      string `json:"so_doc_no"`
}

func (SalesOrderPickActDocView) TableName() string {
	return "vw_get_sales_order_pick_act_doc"
}
