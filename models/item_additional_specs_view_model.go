package models

type AdditionalSpecsView struct {
	AdditionalSpecs
	VolumeUnitOfMeasure string `json:"volume_unit_of_measure"`
	WeightUnitOfMeasure string `json:"weight_unit_of_measure"`
}

func (AdditionalSpecsView) TableName() string {
	return "vw_item_additional_specs"
}
