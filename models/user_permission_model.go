package models

type UserPermissionContent struct {
	ID        uint `gorm:"primarykey; autoIncrement" json:"id"`
	UserId    uint `json:"user_id"`
	CanCreate bool `json:"can_create"`
	CanUpdate bool `json:"can_update"`
	CanDelete bool `json:"can_delete"`
}

type UserPermissionModel struct {
	UserPermissionContent
	User *User `gorm:"foreignKey:UserId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"-"`
}

func (UserPermissionModel) TableName() string {
	return "tbl_user_permission"
}

type UserPermissionAt struct {
	ID    uint `gorm:"primarykey; autoIncrement" json:"id"`
	RefId uint `json:"ref_id"`
	At
	UserPermissionContent
}

func (UserPermissionAt) TableName() string {
	return "z_tbl_user_permission_at"
}
