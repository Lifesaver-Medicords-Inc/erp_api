package models

type PumpTypeContent struct {
	Name string `json:"name"`
}
type PumpType struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	PumpTypeContent
}

func (PumpType) TableName() string {
	return "tbl_setup_item_pump_type"
}

type PumpTypeAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PumpTypeContent
	At
}

func (PumpTypeAt) TableName() string {
	return "z_tbl_setup_item_pump_type_at"
}
