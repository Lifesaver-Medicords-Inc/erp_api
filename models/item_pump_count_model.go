package models

type PumpCountContent struct {
	Name string `json:"name"`
}
type PumpCount struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	PumpCountContent
}
  
func (PumpCount) TableName() string {
	return "tbl_setup_item_pump_count"
}

type PumpCountAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PumpCountContent
	At
}

func (PumpCountAt) TableName() string {
	return "z_tbl_setup_item_pump_count_at"
}
