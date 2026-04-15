package models

type AdditionalSpecsContent struct {
	BasedId                  uint    `json:"based_id"`
	MaterialId               uint    `json:"material_id"`
	SuctionPressure          string  `json:"suction_pressure"`
	DriverType               string  `json:"driver_type"`
	MotorEnclosure           string  `json:"motor_enclosure"`
	MotorManufacturer        string  `json:"motor_manufacturer"`
	ServiceFactor            string  `json:"service_factor"`
	LiquidType               string  `json:"liquid_type"`
	ConnectionType           string  `json:"connection_type"`
	PumpCountCompatabilityId uint    `json:"pump_count_compatibility_id"`
	Size                     string  `json:"size"`
	Volume                   float64 `json:"volume"`
	VolumeUnitOfMeasureId    uint    `json:"volume_unit_of_measure_id"`
	Weight                   float64 `json:"weight"`
	WeightUnitOfMeasureId    uint    `json:"weight_unit_of_measure_id"`
	Calibration              string  `json:"calibration"`
	LongDescription          string  `json:"long_description"`
}

type AdditionalSpecs struct {
	ID uint `gorm:"primarykey" json:"id"`
	AdditionalSpecsContent
}

func (AdditionalSpecs) TableName() string {
	return "tbl_setup_item_additional_specs"
}

type AdditionalSpecsSchema struct {
	AdditionalSpecs
	PumpTypeCompatabilityId []uint `json:"pump_type_compatibility_id"`
}

type AdditionalSpecsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	AdditionalSpecsContent
	At
}

func (AdditionalSpecsAt) TableName() string {
	return "z_tbl_setup_item_additional_specs_at"
}
