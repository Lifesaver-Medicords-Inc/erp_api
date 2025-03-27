package models

type SalesProjectItemsContent struct {
	TemplateID uint `json:"template_id"`
	BasedId    uint `json:"based_id"`
	BomID      uint `json:"bom_id"`
	ItemID     uint `json:"item_id"`

	NodeID       uint   `json:"node_id"`
	NodeName     string `json:"node_name"`
	ParentNodeID uint   `json:"parent_node_id"`
	NodeOrder    uint   `json:"node_order"`
	NodeType     string `json:"node_type"`

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
	ItemsID uint `gorm:"primarykey" json:"items_id"`
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
