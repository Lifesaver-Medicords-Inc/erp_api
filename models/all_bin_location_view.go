package models

type AllBinLocationView struct {
	Location    string `json:"location"`
	WarehouseId uint   `json:"warehouse_id"`
}

func (AllBinLocationView) TableName() string {
	return "vw_get_all_bin_location"
}
