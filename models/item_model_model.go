package models

type ModelContent struct {
	BasedId       uint   `json:"based_id"`
	ItemModelName string `json:"item_model_name"`
	CatalogueYear string `json:"catalogue_year"`
	IsActive      bool   `json:"is_active"`
}

type Model struct {
	ID uint `gorm:"primarykey" json:"id"`
	ModelContent
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
