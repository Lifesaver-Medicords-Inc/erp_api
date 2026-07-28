package models

type OrderDetailsContent struct {
	Based_ID           uint    `json:"based_id"`
	Quotation_Quick_ID uint    `json:"quotation_quick_id"`
	Item_ID            uint    `json:"item_id"`
	DeliveryPreference string  `json:"delivery_preference"`
	Status             string  `json:"status"`
	HasStocks          *bool   `json:"has_stocks"`
	Qty                *int    `json:"qty"`
	Numbering          string  `json:"numbering"`
	ItemCode           string  `json:"item_code"`
	ItemDescription    string  `json:"item_description"`
	ListPrice          float64 `json:"list_price"`
	PercentDiscount    float32 `json:"percent_discount"`
	TotalPrice         float64 `json:"total_price"`
	AllocatedQty       *int    `json:"allocated_qty"`
	OrderType          string  `json:"order_type"`
	BomId              uint    `json:"bom_id"`
	// Dynamic label of the itemset "tab" (from the source project quotation, e.g.
	// tbl_trans_sales_project_item_set.tab_number) that this line item belongs to.
	// Populated only for orders converted from a project quotation; blank/empty for
	// everything else. Kept on each real item row (rather than saving the itemset
	// header itself as its own row) because header rows carry item_id = 0, which
	// violates the item_id foreign key on this table - see the skip in Orders.cs's
	// save handler. Printing reconstructs the header rows from this column by
	// grouping consecutive rows that share the same value.
	ItemSetHeader      string  `json:"item_set_header"`
	Item               *Item   `json:"item"`
	Order              *Order  `gorm:"foreignKey:Based_ID;references:Order_ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
}

type OrderDetails struct {
	OrderDetailsID uint `gorm:"primarykey" json:"order_details_id"`
	OrderDetailsContent
}

func (OrderDetails) TableName() string {
	return "tbl_trans_sales_order_details"
}

type OrderDetailsAt struct {
	OrderDetailsID uint `gorm:"primarykey" json:"order_details_id"`
	RefId          uint `json:"ref_id"`
	OrderDetailsContent
	At
}

func (OrderDetailsAt) TableName() string {
	return "z_tbl_trans_sales_order_details_at"
}
