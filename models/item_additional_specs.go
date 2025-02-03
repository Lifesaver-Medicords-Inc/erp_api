package models

type AdditionalSpecsContent struct {
	BasedId           uint    `json:"based_id"`
	ItemCode          string  `json:"item_code"`
	SuctionPressure   string  `json:"suction_pressure"`
	DriverType        string  `json:"driver_type"`
	MotorEnclosure    string  `json:"motor_enclosure"`
	MotorManufacturer string  `json:"motor_manufacturer"`
	ServiceFactor     string  `json:"service_factor"`
	LiquidType        string  `json:"liquid_type"`
	Volume            float64 `json:"volume"`
	Weight            float64 `json:"weight"`
	LongDescription   string  `json:"long_desc"`
}

type AdditionalSpecs struct {
	ID uint `gorm:"primarykey" json:"id"`
	AdditionalSpecsContent
}

func (AdditionalSpecs) TableName() string {	
	return "tbl_setup_item_additional_specs"
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
