package models

type AdditionalSpecsPumpTypeContent struct {
	AdditionalSpecsId       uint `json:"additional_specs_id"`
	PumpTypeCompatabilityId uint `json:"pump_type_compatability_id"`
}

type AdditionalSpecsPumpType struct {
	ID uint `gorm:"primarykey" json:"id"`
	AdditionalSpecsPumpTypeContent
}

func (AdditionalSpecsPumpType) TableName() string {
	return "tbl_setup_item_additional_specs_pump_type"
}

type AdditionalSpecsPumpTypeAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	AdditionalSpecsPumpTypeContent
	At
}

func (AdditionalSpecsPumpTypeAt) TableName() string {
	return "z_tbl_setup_item_additional_specs_pump_type_at"
}
