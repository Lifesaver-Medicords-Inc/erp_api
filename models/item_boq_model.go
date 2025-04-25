package models

type ItemBoqContent struct {
	ItemID       uint   `json:"item_id"`
	CustomerName string `json:"customer_name"`
	ProjectName  string `json:"project_name"`
	ItemSetName  string `json:"item_set_name"`
	DocRef       string `json:"doc_ref"`
}

type ItemBoq struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemBoqContent
}

func (ItemBoq) TableName() string {
	return "tbl_setup_item_boq"
}

type ItemBoqAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemBoqContent
	At
}

func (ItemBoqAt) TableName() string {
	return "z_tbl_setup_item_boq_at"
}
