package models

type SalesProjectItemSetContent struct {
	BasedId   uint   `json:"based_id"`
	TabNumber string `json:"tab_number"`
}

type SalesProjectItemSet struct {
	ItemSetID uint `gorm:"primaryKey;column:itemset_id" json:"itemset_id"`
	SalesProjectItemSetContent
}

func (SalesProjectItemSet) TableName() string {
	return "tbl_trans_sales_project_item_set"
}

type SalesProjectItemSetAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefID uint `json:"ref_id"`
	SalesProjectItemSetContent
	At
}

func (SalesProjectItemSetAt) TableName() string {
	return "z_tbl_trans_sales_project_item_set_at"
}
