package models

type User struct {
	ID         uint   `gorm:"primarykey"`
	EmployeeId string `gorm:"unique"`
	Password   string
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
