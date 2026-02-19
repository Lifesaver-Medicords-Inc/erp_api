package models

type ItemInventoryContent struct {
	BasedId            uint    `json:"based_id"`
	WarehouseId        uint    `json:"warehouse_id"`
	DefaultZone        string  `json:"default_zone"`
	StorageType        string  `json:"storage_type"`
	DefaultBinLocation string  `json:"default_bin_location"`
	ValuationMethodId  uint    `json:"valuation_method_id"`
	MinimunInventory   float64 `json:"minimum_inventory"`
	MaximunInventory   float64 `json:"maximum_inventory"`
	IsSpecialItem      *bool   `json:"is_special_item"`
}
type ItemInventory struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemInventoryContent
}

func (ItemInventory) TableName() string {
	return "tbl_setup_item_inventory"
}

type ItemInventoryAt struct {
	ID    uint `gorm:"primaryKey;autoIncrement" json:"id"`
	RefId uint `json:"ref_id"`
	ItemInventoryContent
	At
}

func (ItemInventoryAt) TableName() string {
	return "z_tbl_setup_item_inventory_at"
}
