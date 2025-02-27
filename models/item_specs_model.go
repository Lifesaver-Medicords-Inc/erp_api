package models

type ItemSpecsContent struct {
	BasedId            uint   `json:"based_id"`
	Template           string `json:"template"`
	Title              string `json:"title"`
	Value              string `json:"value"`
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
