package models

type ItemContent struct {
	ItemNameId          uint    `json:"item_name_id"`
	ItemModel           string  `json:"item_model"`
	CatalogueYear       string  `json:"catalogue_year"`
	ItemCode            string  `json:"item_code"`
	ShortDesc           string  `json:"short_desc"`
	ItemClassId         uint    `json:"item_class_id"`
	ItemBrandId         uint    `json:"item_brand_id"`
	UnitOfMeasureId     uint    `json:"unit_of_measure_id"`
	ItemTangibilityType string  `json:"item_tangibility_type"`
	IsStopSelling       *int   `json:"is_stop_selling"`
	Price               float64 `json:"price"`
}
type Item struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemContent
}

func (Item) TableName() string {
	return "tbl_setup_item"
}

type ItemAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`

	ItemContent
	At
}

func (ItemAt) TableName() string {
	return "z_tbl_setup_item_at"
}
