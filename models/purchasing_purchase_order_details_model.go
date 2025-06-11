package models

type PurchaseOrderDetailsContent struct {
	BasedId         uint    `json:"based_id"`
	ItemId          int     `json:"item_id"`
	ItemCode        string  `json:"item_code"`
	ItemName        string  `json:"item_name"`
	ReqQty          int     `json:"req_qty"`
	OrderQty        int     `json:"order_qty"`
	UnitOfMeasure   string  `json:"unit_of_measure"`
	UnitPrice       float64 `json:"unit_price"`
	Discount        string  `json:"discount"`
	DiscountedPrice float64 `json:"discounted_price"`
	TotalPrice      float64 `json:"total_price"`
	OrderDetailIds  string  `json:"order_detail_ids"`
	AllocatedQtys   string  `json:"allocated_qtys"`
	Qtys            string  `json:"qtys"`
}

type PurchaseOrderDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	PurchaseOrderDetailsContent
}

func (PurchaseOrderDetails) TableName() string {
	return "tbl_purchasing_purchase_order_details"
}

type PurchaseOrderDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchaseOrderDetailsContent
	At
}

func (PurchaseOrderDetailsContent) TableName() string {
	return "z_tbl_purchasing_purchase_order_details_at"
}
