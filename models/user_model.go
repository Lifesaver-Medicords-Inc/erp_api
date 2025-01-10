package models

type User struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
	EmployeeId string `gorm:"unique" json:"employee_id"`
	Password   string `json:"password"`
}

func (User) TableName() string {
	return "tbl_setup_users"
}

type UserAt struct {
	User
	At
}

func (UserAt) TableName() string {
	return "z_tbl_setup_users_at"
}
