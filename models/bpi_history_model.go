package models

type BpiHistoryView struct {
	BasedId     uint   `json:"based_id"`
	BranchId    uint   `json:"branch_id"`
	AtDate      string `json:"at_date"`
	Actions     string `json:"actions"`
	EditBy      string `json:"edit_by"`
	EditHistory string `json:"edit_history"`
}

func (BpiHistoryView) TableName() string {
	return "vw_get_bpi_history"
}
