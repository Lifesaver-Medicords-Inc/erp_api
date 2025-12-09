package models

type BpiHistoryDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiHistoryDetailsContent
}

type BpiHistoryDetailsContent struct {
	BasedId  uint   `json:"based_id"`
	BranchId uint   `json:"branch_id"`
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

func (BpiHistoryDetails) TableName() string {
	return "tbl_bpi_history_details"
}

type BpiHistoryDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiHistoryDetailsContent
	At
}

func (BpiHistoryDetailsAt) TableName() string {
	return "z_tbl_bpi_history_details_at"
}
