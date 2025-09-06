package models

type UserContent struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Password   string `json:"password"`
	PositionId uint   `json:"position_id"`
	Department string `json:"department"`
}

type User struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	EmployeeId string `gorm:"unique" json:"employee_id"`
	UserContent
	Position    Position        `gorm:"foreignKey:PositionId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"position"`
	Permissions *UserPermission `gorm:"foreignKey:UserId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"permissions"`
}

func (User) TableName() string {
	return "tbl_setup_users"
}

type UserAt struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	RefId      uint   `json:"ref_id"`
	EmployeeId string `json:"employee_id"`
	UserContent
	At
}

func (UserAt) TableName() string {
	return "z_tbl_setup_users_at"
}
