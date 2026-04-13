package inventory_models

type WarehouseReceivingView struct {
	WarehouseId      uint   `json:"warehouse_id"`
	Warehouse        string `json:"warehouse"`
	WarehouseAddress string `json:"warehouse_address"`
}

func (WarehouseReceivingView) TableName() string {
	return "vw_get_warehouse_receiving"
}

type WarehouseReceivingAreaView struct {
	WarehouseAreaId uint   `json:"warehouse_area_id"`
	Zone            string `json:"zone"`
	Area            string `json:"area"`
	Rack            string `json:"rack"`
	Level           string `json:"level"`
	Bins            string `json:"bins"`
}
