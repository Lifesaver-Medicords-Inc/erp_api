package models

type MaterialContent struct {
	Name string `json:"name"`
}

type Material struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	MaterialContent
}

func (Material) TableName() string {
	return "tbl_setup_item_material"
}

type MaterialAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	MaterialContent
	At
}

func (MaterialAt) TableName() string {
	return "z_tbl_setup_item_material_at"
}
