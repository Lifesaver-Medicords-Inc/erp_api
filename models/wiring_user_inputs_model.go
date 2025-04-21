package models

type WiringUserInputContent struct {
	WiringID   uint   `json:"wiring_id"` // References SalesProjectWiring.id
	WiringNote string `json:"wiring_note"`
}

type WiringUserInput struct {
	ID uint `gorm:"primarykey" json:"id"`
	WiringUserInputContent
}

func (WiringUserInput) TableName() string {
	return "tbl_wiring_user_inputs"
}

type WiringUserInputAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	WiringUserInputContent
	At
}

func (WiringUserInputAt) TableName() string {
	return "z_tbl_wiring_user_inputs_at"
}
