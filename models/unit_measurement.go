package models

type UnitMeasurementContent struct {
	Name string `json:"name"`
}

type UnitMeasurement struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	UnitMeasurementContent
}

func (UnitMeasurement) TableName() string {
	return "tbl_setup_item_unit_measurement"
}

type UnitMeasurementAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	UnitMeasurementContent
	At
}

func (UnitMeasurementAt) TableName() string {
	return "z_tbl_setup_item_unit_measurement_at"
}
