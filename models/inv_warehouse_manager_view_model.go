package models

type WarehouseManagerContent struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

type WarehouseManager struct {
	WarehouseManagerContent
}

func (WarehouseManager) TableName() string {
	return "vw_warehouse_manager"
}
