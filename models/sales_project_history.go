package models

type SalesProjectHistoryContent struct {
	BasedId uint   `json:"based_id"`
	User    string `json:"user"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	OldData string `json:"old_data"`
	NewData string `json:"new_data"`
}

type SalesProjectHistory struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesProjectHistoryContent
}

func (SalesProjectHistory) TableName() string {
	return "tbl_trans_sales_project_history"
}

type SalesProjectHistoryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesProjectHistoryContent
	At
}

func (SalesProjectHistoryAt) TableName() string {
	return "z_tbl_trans_sales_project_history_at"
}
