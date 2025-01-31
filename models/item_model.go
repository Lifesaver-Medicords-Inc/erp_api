package models

type ItemContent struct {
	ItemNameId        uint   `json:"item_name_id"`
	ItemModelId       uint   `json:"item_model_id"`
	ItemCode          string `json:"item_code"`
	ShortDesc         string `json:"short_desc"`
	ItemClassId       uint   `json:"item_class_id"`
	ItemBrandId       uint   `json:"item_brand_id"`
	UnitOfMeasureId   uint   `json:"unit_of_measure_id"`
	IsStopSelling     bool   `json:"is_stop_selling"`
	IsSalesItem       bool   `json:"is_sales_item"`
	IsPurchaseItem    bool   `json:"is_purchase_item"`
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
