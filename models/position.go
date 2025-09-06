package models

type PositionContent struct {
	Name string `json:"name"`
}

type Position struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	PositionContent
	Users  []User           `gorm:"foreignKey:PositionId" json:"users"`
	Access []PositionAccess `gorm:"foreignKey:PositionId" json:"access"`
}

func (Position) TableName() string {
	return "tbl_position"
}

type PositionAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PositionContent
	At
}

func (PositionAt) TableName() string {
	return "z_tbl_position_at"
}
