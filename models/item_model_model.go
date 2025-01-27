package models

type ModelContent struct {
	Name          string `json:"name"`
	ItemNameId    int    `json:"item_name_id"`
	ItemBrandId   int    `json:"item_brand_id"`
	CatalogueYear string `json:"catalogue_year"`
	IsActive      bool   `json:"is_active"`
}
type Model struct {
	ID uint `gorm:"primarykey" json:"id"`
	ModelContent
	RelatedItemId string `json:"related_item_id" gorm:"->"`
	RelatedName string `json:"related_name" gorm:"->"`
	RelatedBrand  string `json:"related_brand" gorm:"->"`
}

func (Model) TableName() string {
	return "tbl_setup_item_model"
}

type ModelAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ModelContent
	At
}

func (ModelAt) TableName() string {
	return "z_tbl_setup_item_model_at"
}
