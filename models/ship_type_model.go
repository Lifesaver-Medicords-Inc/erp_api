package models

type ShipTypeContent struct {
	ShipName string `json:"shipname"`
}

type ShipType struct {
	ID uint `gorm:"primarykey" json:"id"`
	ShipTypeContent
}

func (ShipType) TableName() string {
	return "tbl_setup_ship_type"
}

type ShipTypeAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ShipTypeContent
	At
}

func (ShipTypeAt) TableName() string {
	return "z_tbl_setup_ship_type_at"
}
