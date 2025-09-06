package models

type VehicleContent struct {
	Type            string `json:"type"`
	Model           string `json:"model"`
	Description     string `json:"Description"`
	PlateNo         string `json:"plate_no"`
	AcquisitionYear string `json:"acquisition_year"`
	Capacity        uint   `json:"capacity"`
	Status          string `json:"Status"`
	LastMaintenance string `json:"last_maintenance"`
	Notes           string `json:"notes"`
}

type Vehicle struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	WareHouseId int  `json:"warehouse_id"`
	VehicleContent
}

func (Vehicle) TableName() string {
	return "tbl_vehicle"
}

type VehicleAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	VehicleContent
	At
}

func (VehicleAt) TableName() string {
	return "z_tbl_vehicle_at"
}
