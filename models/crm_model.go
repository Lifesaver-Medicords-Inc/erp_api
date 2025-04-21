package models

type CRMContent struct {
	BasedID uint   `json:"based_id"`
	TAG     string `json:"tag"`
	Date    string `json:"date"`
	Remark  string `json:"remark"`
}
type CRM struct {
	CRM_ID uint `gorm:"primarykey" json:"crm_id"`
	CRMContent
}

// test
func (CRM) TableName() string {
	return "tbl_trans_sales_crm"
}

type CRMAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	CRMContent
	At
}

func (CRMAt) TableName() string {
	return "z_tbl_trans_sales_crm_at"
}
