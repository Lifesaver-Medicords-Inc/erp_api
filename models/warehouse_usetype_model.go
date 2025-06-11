package models

type WarehouseUseTypeContent struct {
	Name string `json:"name"`
}

type WarehouseUseType struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;not null" json:"code"`
	WarehouseUseTypeContent
}

func (WarehouseUseType) TableName() string {
	return "tbl_setup_usetype"
}

type WarehouseUseTypeAt struct {
	ID    uint   `gorm:"primary" json:"id"`
	RefID uint   `json:"ref_id"`
	Code  string `json:"code"`
	WarehouseUseTypeContent
	At
}

func (WarehouseUseTypeAt) TableName() string {
	return "z_tbl_setup_usetype_at"
}
