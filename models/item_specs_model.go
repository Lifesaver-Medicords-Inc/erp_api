package models

type ItemSpecsContent struct {
	BasedId  uint   `json:"based_id"`
	ItemCode string `json:"item_code"`
	Template string `json:"template"`
	Title    string `json:"title"`
	Value    string `json:"value"`
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

func (ItemSpecsContent) TableName() string {
	return "z_tbl_setup_item_specs_at"
}
