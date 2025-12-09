package models

type BpiItemListContent struct {
	ItemType            string  `json:"item_type"`
	ItemCode        string  `json:"item_code"`
	GeneralName     string  `json:"general_name"`
	ItemModelName   string  `json:"item_model_name"`
	ItemBrandName   string  `json:"item_brand_name"`
	LongDescription string  `json:"long_description"`
	ItemPrice       float64 `json:"item_price"`
}

type BpiItemList struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiItemListContent
}

func (BpiItemList) TableName() string {
	return "vw_bpi_item_list"
}
