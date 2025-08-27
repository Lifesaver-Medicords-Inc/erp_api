package adminmodels

import "github.com/pierceperado/smpc/models"

type UserPermission struct {
	ID        uint `gorm:"primarykey; autoIncrement" json:"id"`
	UserId    uint `json:"user_id"`
	CanCreate bool `json:"can_create"`
	CanUpdate bool `json:"can_update"`
	CanDelete bool `json:"can_delete"`
}

func (UserPermission) TableName() string {
	return "tbl_admin_user_permission"
}

type UserPermissionAt struct {
	ID    uint `gorm:"primarykey; autoIncrement" json:"id"`
	RefId uint `json:"ref_id"`
	models.At
	UserPermission
}

func (UserPermissionAt) TableName() string {
	return "tbl_admin_user_permission_at"
}
