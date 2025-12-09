package models

type BpiHistory struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiHistoryContent
}
type BpiHistoryContent struct {
	BasedId   uint   `json:"based_id"`
	Date      string `json:"date"`
	Actions   string `json:"actions"`
	ChildType string `json:"child_type"`
	EditBy    string `json:"edit_by"`
}

func (BpiHistory) TableName() string {
	return "tbl_bpi_history"
}

type BpiHistoryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiHistoryContent
	At
}

func (BpiHistoryAt) TableName() string {
	return "z_tbl_bpi_history_at"
}
