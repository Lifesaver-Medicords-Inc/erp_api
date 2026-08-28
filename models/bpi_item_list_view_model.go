package models

type BpiItemListContent struct {
	ItemType        string  `json:"item_type"`
	ItemCode        string  `json:"item_code"`
	GeneralName     string  `json:"general_name"`
	ItemModelName   string  `json:"item_model_name"`
	ItemBrandName   string  `json:"item_brand_name"`
	LongDescription string  `json:"long_description"`
	ItemPrice       float64 `json:"item_price"`
	// ShortDesc/StatusTangible/StatusTrade added to fix the BPI Item picker
	// (BusinessPartnerInfo.cs's dg_items_CellClick) reading dictionary keys
	// GetResult() never provided - it threw KeyNotFoundException the moment
	// that path ran. Additive only: ItemType/LongDescription stay as-is since
	// BpiBranchUC.cs's own picker handler already reads those two correctly.
	ShortDesc      string `json:"short_desc"`
	StatusTangible string `json:"status_tangible"`
	StatusTrade    string `json:"status_trade"`
}

type BpiItemList struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiItemListContent
}

func (BpiItemList) TableName() string {
	return "vw_bpi_item_list"
}
