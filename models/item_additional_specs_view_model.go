package models

type AdditionalSpecsView struct {
	AdditionalSpecs
	VolumeUnitOfMeasure        string `json:"volume_unit_of_measure"`
	WeightUnitOfMeasure        string `json:"weight_unit_of_measure"`
	LengthUnitOfMeasure        string `json:"length_unit_of_measure"`
	HeightUnitOfMeasure        string `json:"height_unit_of_measure"`
	PumpTypeCompatabilityId    string `json:"pump_type_compatability_id"`
	PumpTypeCompatabilityNames string `json:"pump_type_compatability_names"`
}

func (AdditionalSpecsView) TableName() string {
	return "vw_item_additional_specs"
}
