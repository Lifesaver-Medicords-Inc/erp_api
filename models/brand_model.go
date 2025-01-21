package models

type BrandContent struct {
	Name string `json:"name"`
}

type Brand struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	BrandContent
}

func (Brand) TableName() string {
	return "tbl_setup_item_brand"
}

type BrandAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	BrandContent
	At
}

func (BrandAt) TableName() string {
	return "z_tbl_setup_item_brand_at"
}
