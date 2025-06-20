package models

type WarehouseNameContent struct {
	Name             string `json:"name"`
	WarehouseManager string `json:"warehouse_manager"`
	IsInactive       *bool  `json:"is_inactive"`
}

type WarehouseName struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	WarehouseNameContent
}

func (WarehouseName) TableName() string {
	return "tbl_warehouse_name"
}

type WarehouseNameAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	WarehouseNameContent

	At
}

func (WarehouseNameAt) TableName() string {
	return "z_tbl_warehouse_name_at"
}
