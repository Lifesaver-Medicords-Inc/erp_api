package models

type PositionAccessContent struct {
	ID         uint          `gorm:"primarykey; autoIncrement" json:"id"`
	PositionId uint          `json:"position_id"`
	Code       string        `gorm:"not null" json:"code"`
	Position   PositionModel `gorm:"foreignKey:PositionId;references:ID" json:"position"`
}

type PositionAccessModel struct {
	PositionAccessContent
}

type PositionAccessAt struct {
	ID    uint   `gorm:"primarykey; autoIncrement" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	At
	PositionAccessContent
}

func (PositionAccessModel) TableName() string {
	return "tbl_position_access"
}

func (PositionAccessAt) TableName() string {
	return "z_tbl_position_access_at"
}
