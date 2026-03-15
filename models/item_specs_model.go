package models

type ItemSpecsContent struct {
	BasedId            uint   `json:"based_id"`
	Template           string `json:"template"`
	Title              string `json:"title"`
	Value              string `json:"value"`
	Fla_1              string `json:"fla_1"`
	Fla_2              string `json:"fla_2"`
	Volt_1             string `json:"volt_1"`
	Volt_2             string `json:"volt_2"`
	ManufacturerOrigin string `json:"manufacturer_origin"`
}

type ItemSpecs struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemSpecsContent
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
