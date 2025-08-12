package models

type PROrdersContent struct {
	Based_ID        uint   `json:"based_id"`
	ItemID          uint   `json:"item_id"`
	QTY             uint   `json:"qty"`
	Status          string `json:"status"`
	ItemCode        string `json:"item_code"`
	ItemDescription string `json:"item_description"`
	AllocatedQty    *int   `json:"allocated_qty"`
	OrderType       string `json:"order_type"`
}

type PROrders struct {
	PR_Order_ID uint `gorm:"primarykey" json:"pr_order_id"`
	PROrdersContent
}

func (PROrders) TableName() string {
	return "tbl_purchasing_purchase_requisition_orders"
}

type PROrdersAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PROrdersContent
	At
}

func (PROrdersAt) TableName() string {
	return "z_tbl_purchasing_purchase_requisition_orders_at"
}
