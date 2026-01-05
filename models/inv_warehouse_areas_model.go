package models

type WarehouseAreaContent struct {
	WarehouseNameId uint   `json:"warehouse_name_id"` //parent id
	UseType         string `json:"use_type"`
	Zone            string `json:"zone"`
	Area            string `json:"area"`
	Rack            string `json:"rack"`
	Level           string `json:"level"`
	Bins            string `json:"bins"`
	LocationCode    string `json:"location_code"`
	Notes           string `json:"notes"`
}

type WarehouseArea struct {
	ID uint `gorm:"primarykey" json:"id"`
	WarehouseAreaContent
}

func (WarehouseArea) TableName() string {
	return "tbl_inv_warehouse_area"
}

type WarehouseAreaAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	WarehouseAreaContent
	At
}

func (WarehouseAreaAt) TableName() string {
	return "z_tbl_inv_warehouse_area_at"
}

type WarehouseAreaView struct {
	WarehouseNameId uint   `json:"warehouse_name_id"` //parent id
	Zone            string `json:"zone"`
	Area            string `json:"area"`
	Rack            string `json:"rack"`
	Level           string `json:"level"`
	Bins            string `json:"bins"`
	LocationCode    string `json:"location_code"`
	WarehouseName   string `json:"warehouse_name"`
}

func (WarehouseAreaView) TableName() string {
	return "vw_get_bin_loc_pick_activity"
}
