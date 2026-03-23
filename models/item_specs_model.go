package models

type ItemSpecsContent struct {
	BasedId            uint   `json:"based_id"`
	Template           string `json:"template"`
	Fla_1              string `json:"fla_1"`
	Fla_2              string `json:"fla_2"`
	Volt_1             string `json:"volt_1"`
	Volt_2             string `json:"volt_2"`
	ImpellerId         uint   `json:"impeller_id"`
	ManufacturerOrigin string `json:"manufacturer_origin"`
}

type ItemSpecs struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemSpecsContent
	ItemSpecsTemplate []ItemSpecsTemplate `gorm:"foreignKey:BasedId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item_specs_template,omitempty"`
}

func (ItemSpecs) TableName() string {
	return "tbl_setup_item_specs"
}

type ItemSpecsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemSpecsContent
	At
}

func (ItemSpecsAt) TableName() string {
	return "z_tbl_setup_item_specs_at"
}
