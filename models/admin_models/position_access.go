package adminmodels

import "github.com/pierceperado/smpc/models"

type PositionAccess struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	PositionId uint   `json:"position_id"`
	Code       string `gorm:"not null" json:"code"`
}

type PositionAccessAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	models.At
	PositionAccess
}

func (PositionAccess) TableName() string {
	return "tbl_admin_position_access"
}

func (PositionAccessAt) TableName() string {
	return "tbl_admin_position_access_at"
}
