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
