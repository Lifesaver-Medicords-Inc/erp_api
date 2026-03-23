package models

type ItemSpecsTemplateContent struct {
	BasedId uint   `json:"based_id"`
	Title   string `json:"title"`
	Value   string `json:"value"`
}

type ItemSpecsTemplate struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemSpecsTemplateContent
}

func (ItemSpecsTemplate) TableName() string {
	return "tbl_setup_item_specs_template"
}

type ItemSpecsTemplateAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemSpecsTemplateContent
	At
}

func (ItemSpecsTemplateAt) TableName() string {
	return "z_tbl_setup_item_specs_template_at"
}
