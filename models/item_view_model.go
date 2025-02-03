package models

type ItemView struct {
	Item
	ItemName      string `json:"item_name"`
	ItemModel     string `json:"item_model"`
	ItemClass     string `json:"item_class"`
	ItemBrand     string `json:"item_brand"`
	UnitOfMeasure string `json:"unit_of_measure"`
}

func (ItemView) TableName() string {
	return "ItemList"
}
