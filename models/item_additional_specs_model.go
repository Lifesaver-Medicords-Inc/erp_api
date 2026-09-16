package models

import "encoding/json"

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

	// Whether the request carried pump_count_compatibility_id / pump_type_compatibility_id at
	// all - set by UnmarshalJSON, never serialized. An update writes those two only when they
	// were sent: a missing key and a real "none" both decode to zero, and apps built before
	// the key spelling was fixed send neither.
	PumpCountSent bool `json:"-"`
	PumpTypesSent bool `json:"-"`
}

// UnmarshalJSON decodes as usual, then records which of the two pump keys were present.
func (s *AdditionalSpecsSchema) UnmarshalJSON(data []byte) error {
	type schema AdditionalSpecsSchema // a defined type has none of these methods, so no recursion
	if err := json.Unmarshal(data, (*schema)(s)); err != nil {
		return err
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	_, s.PumpCountSent = keys["pump_count_compatibility_id"]
	_, s.PumpTypesSent = keys["pump_type_compatibility_id"]
	return nil
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
