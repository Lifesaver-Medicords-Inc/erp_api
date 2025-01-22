package models

type Brand struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
}

func (Brand) TableName() string {
	return "tbl_setup_item_brand"
}

type BrandAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At    `json:"at"`
}

func (BrandAt) TableName() string {
	return "z_tbl_setup_item_brand_at"
}
