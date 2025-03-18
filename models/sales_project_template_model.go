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
	NodeID    uint   `json:"node_id"`
	BasedID   uint   `json:"based_id"`
	NodeName  string `json:"node_name"`
	NodeLevel uint   `json:"node_level"`
	NodeOrder uint   `json:"node_order"`
	ItemID    uint   `json:"item_id"`
	NodeType  string `json:"node_type"`
}

type SalesProjectTemplateChild struct {
	NodesID      uint `gorm:"primarykey" json:"nodes_id"`
	ParentNodeID uint `json:"parent_node_id"`
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
