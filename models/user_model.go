package models

type UserContent struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
	Password   string `json:"password"`
	PositionId uint	  `json:"position_id"`
}

type User struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	EmployeeId string `gorm:"unique" json:"employee_id"`
	UserContent
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
	User User `gorm:"foreignKey:RefId;references:ID;onDelete:CASCADE;onUpdate:CASCADE"`
}

func (UserAt) TableName() string {
	return "z_tbl_setup_users_at"
}
