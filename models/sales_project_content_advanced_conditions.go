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

	// The six Advanced Conditions dropdowns, held as IDs rather than as their text
	// (user decision 2026-09-24, the same rule as flow_id/head_id on the content row).
	//
	// PumpBrandId points at tbl_setup_item_brand - the list Item Entry uses - so a brand
	// renamed in setup reads correctly on every quotation already saved. The other five
	// are lists that exist only in the client for now; storing the id is what lets any of
	// them become a Setup list later without rewriting stored quotations:
	//
	//   driver_type_id             1 Electric Motor, 2 Engine
	//   motor_enclosure_id         1 NEMA 2, 2 NEMA 3, 3 NEMA 4, 4 NEMA 12
	//   motor_manufacturer_id      1 Teco Motor, 2 Franklin Motor, 3 Calpeda Motor
	//   liquid_type_id             1 Clean, 2 Sewage, 3 Seawater
	//   controller_manufacturer_id 1 Eaton, 2 Marathon, 3 WEG
	//
	// The ids are fixed: append, never renumber. 0 is "-- Select --", which is what an
	// unanswered dropdown stores (spec 1.3) - not the first value.
	//
	// The nvarchar columns above are kept and left alone: they are empty on every row
	// that exists, and dropping a column is not something this needs to do.
	PumpBrandId              uint `json:"pump_brand_id"`
	DriverTypeId             uint `json:"driver_type_id"`
	MotorEnclosureId         uint `json:"motor_enclosure_id"`
	MotorManufacturerId      uint `json:"motor_manufacturer_id"`
	LiquidTypeId             uint `json:"liquid_type_id"`
	ControllerManufacturerId uint `json:"controller_manufacturer_id"`
}

type SalesProjectAdvancedConditions struct {
	ConditionsID uint `json:"conditions_id" gorm:"primaryKey"`
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
