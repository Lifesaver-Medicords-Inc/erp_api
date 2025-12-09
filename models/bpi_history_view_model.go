package models

type BpiHistoryView struct {
	BranchId    uint   `json:"branch_id"`
	Date        string `json:"date"`
	Actions     string `json:"actions"`
	EditBy      string `json:"edit_by"`
	EditHistory string `json:"edit_history"`
}

func (BpiHistoryView) TableName() string {
	return "vw_get_bpi_history"
}
