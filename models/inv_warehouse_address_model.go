package models

type WarehouseAddressContent struct {
	WarehouseNameId uint   `json:"warehouse_name_id"` // parent id
	BuildingNo      string `json:"building_no"`
	Street          string `json:"street"`
	BarangayNo      string `json:"barangay_no"`
	City            string `json:"city"`
	ZipCode         string `json:"zip_code"`
	Country         string `json:"country"`
	ContactPerson   string `json:"contact_person"`
	ContactNo       string `json:"contact_no"`
}

type WarehouseAddress struct {
	ID uint `gorm:"primarykey" json:"id"`
	WarehouseAddressContent
}

func (WarehouseAddress) TableName() string {
	return "tbl_inv_warehouse_address"
}

type WarehouseAddressAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	WarehouseAddressContent
	At
}

func (WarehouseAddressAt) TableName() string {
	return "z_tbl_inv_warehouse_address_at"
}
