package models

type ItemView struct {
	Item
	ItemName      string `json:"item_name" gorm:"->"`
	ItemModel     string `json:"item_model" gorm:"->"`
	ItemClass     string `json:"item_class" gorm:"->"`
	ItemBrand     string `json:"item_brand" gorm:"->"`
	UnitOfMeasure string `json:"unit_of_measure" gorm:"->"`
}

func (ItemView) TableName() string {
	return "ItemList"
}
