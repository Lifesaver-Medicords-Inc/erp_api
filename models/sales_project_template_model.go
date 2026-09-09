package models

type SalesProjectTemplateContent struct {
	TemplateName string `json:"template_name"`
}

type SalesProjectTemplate struct {
	TemplateID uint `gorm:"primarykey" json:"template_id"`
	SalesProjectTemplateContent
}

func (SalesProjectTemplate) TableName() string {
	return "tbl_trans_sales_project_template"
}

type SalesProjectTemplateAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesProjectTemplateContent
	At
}

func (SalesProjectTemplateAt) TableName() string {
	return "z_tbl_trans_sales_project_template_at"
}

type SalesProjectTemplateChildContent struct {
	ParentId   uint   `json:"parent_id"`
	ItemID     uint   `json:"item_id"`
	Components string `json:"components"`
	Level      uint   `json:"level"`

	// Qty - the quantity of this component the template proposes, carried
	// into a project quotation's own qty column when the template is applied.
	//
	// A POINTER, not a plain uint, because "nobody has set a quantity" and
	// "the quantity is zero" have to stay distinguishable. Every one of the
	// component rows that existed before this column did is unset, and the
	// agreed behaviour is that they stay blank and keep applying blank - a
	// plain uint would turn all of them into 0, which displays as "0" on the
	// quotation and reads as a deliberate zero rather than as "not specified".
	//
	// Quantity is deliberately NOT defaulted anywhere - not on a new template
	// row, not on apply. It exists only where somebody typed it, so adding
	// this column changes no existing quotation's total.
	Qty *uint `json:"qty"`
}

type SalesProjectTemplateChild struct {
	ID       uint `gorm:"primarykey" json:"id"`
	ParentID uint `json:"parent_id"`
	SalesProjectTemplateChildContent
}

func (SalesProjectTemplateChild) TableName() string {
	return "tbl_trans_sales_project_template_child"
}

type SalesProjectTemplateChildAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesProjectTemplateChildContent
	At
}

func (SalesProjectTemplateChildAt) TableName() string {
	return "z_tbl_trans_sales_project_template_child_at"
}
