package adminmodels

import "github.com/pierceperado/smpc/models"

type Position struct {
	ID     uint              `gorm:"primarykey; autoIncrement" json:"id"`
	Name   string            `gorm:"unique" json:"name"`
	Access []*PositionAccess `gorm:"foreignKey:PositionId" json:"access"`
}

type PositionAt struct {
	ID    uint `gorm:"primarykey; autoIncrement" json:"id"`
	RefId uint `json:"ref_id"`
	models.At
	Position
}

func (Position) TableName() string {
	return "tbl_admin_position"
}

func (PositionAt) TableName() string {
	return "tbl_admin_position_at"
}
