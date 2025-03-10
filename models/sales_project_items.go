package models

type SalesProjectItemsContent struct {
	BasedId          uint    `json:"based_id"`
	Components       string  `json:"components"`
	Model            string  `json:"model"`
	ItemInvType      string  `json:"item_inv_type"`
	Qty              uint    `json:"qty"`
	Multiplier       string  `json:"multiplier"`
	DiscountPrice    float64 `json:"discount_price"`
	ListPricePerUnit float64 `json:"list_price_per_unit"`
	ComponentTotal   float64 `json:"component_total"`
	Notes            string  `json:"notes"`
}

type SalesProjectItems struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesProjectItemsContent
}

func (SalesProjectItems) TableName() string {
	return "tbl_trans_sales_project_items"
}

type SalesProjectItemsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefID uint `json:"ref_id"`
	SalesProjectItemsContent
	At
}

func (SalesProjectItemsAt) TableName() string {
	return "z_tbl_trans_sales_project_items_at"
}
