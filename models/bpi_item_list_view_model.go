package models

type BpiItemListContent struct {
	ShortDesc      string `json:"short_desc"`
	ItemCode       string `json:"item_code"`
	GeneralName    string `json:"general_name"`
	ItemModelName  string `json:"item_model_name"`
	ItemBrandName  string `json:"item_brand_name"`
	StatusTrade    string `json:"status_trade"`
	StatusTangible string `json:"status_tangible"`
}

type BpiItemList struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiItemListContent
}

func (BpiItemList) TableName() string {
	return "vw_bpi_item_list"
}
