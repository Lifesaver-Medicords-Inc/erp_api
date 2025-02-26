package models

type SalesProjectAdvancedConditionsContent struct {
	// SHOULD BE THE TAB # / SET #
	BasedId                uint   `json:"based_id"`
	PumpBrand              string `json:"pump_brand"`
	DriverType             string `json:"driver_type"`
	Pressure               string `json:"pressure"`
	MotorEnclosure         string `json:"motor_enclosure"`
	MotorManufacturer      string `json:"motor_manufacturer"`
	LiquidType             string `json:"liquid_type"`
	ControllerManufacturer string `json:"controller_manufacturer"`
	StartingMethod         string `json:"starting_method"`
	SuctionSize            string `json:"suction_size"`
	DischargeSize          string `json:"discharge_size"`
}

type SalesProjectAdvancedConditions struct {
	ID uint `json:"id" gorm:"primaryKey"`
	SalesProjectAdvancedConditionsContent
}

func (SalesProjectAdvancedConditions) TableName() string {
	return "tbl_trans_sales_project_content_advanced_configuration"
}

type SalesProjectAdvancedConditionsAt struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	RefID uint `json:"ref_id"`
	SalesProjectAdvancedConditionsContent
	At
}

func (SalesProjectAdvancedConditionsAt) TableName() string {
	return "tbl_trans_sales_project_content_advanced_configuration_at"
}
