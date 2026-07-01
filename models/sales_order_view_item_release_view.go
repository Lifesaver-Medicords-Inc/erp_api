package models

type SalesOrderItemReleaseView struct {
	SalesOrderDetailsID uint   `json:"sales_order_details_id"`
	SalesOrderID        uint   `json:"sales_order_id"`
	RefDocNo            string `json:"ref_doc_no"`
	ItemID              uint   `json:"item_id"`
	ItemCode            string `json:"item_code"`
	ItemDescription     string `json:"item_description"`
	RequiredQty         uint   `json:"required_qty"`
	RequiredUomID       uint   `json:"required_uom_id"`
	RequiredUom         string `json:"required_uom"`
	DeliveryPreference  string `json:"delivery_preference"`
	ReleasedQty         *uint  `json:"released_qty"`
	ReleaseUom          string `json:"release_uom"`
	SerialNo            string `json:"serial_no"`
}

func (SalesOrderItemReleaseView) TableName() string {
	return "vw_get_sales_order_item_release"
}
