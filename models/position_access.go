package models

type PositionAccessContent struct {
	ID         uint     `gorm:"primarykey; autoIncrement" json:"id"`
	PositionId uint     `json:"position_id"`
	Code       string   `gorm:"not null" json:"code"`
	Position   Position `gorm:"foreignKey:PositionId;references:ID" json:"position"`
}

type PositionAccess struct {
	PositionAccessContent
}

type PositionAccessAt struct {
	ID    uint   `gorm:"primarykey; autoIncrement" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	At
	PositionAccessContent
	PositionAccess PositionAccess `gorm:"foreignKey:RefId;references:ID;onDelete:CASCADE;onUpdate:CASCADE"`
}

func (PositionAccess) TableName() string {
	return "tbl_position_access"
}

func (PositionAccessAt) TableName() string {
	return "tbl_position_access_at"
}
