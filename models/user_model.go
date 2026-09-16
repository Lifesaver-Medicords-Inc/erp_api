package models

import "encoding/json"

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
	Position    PositionModel        `gorm:"foreignKey:PositionId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"position"`
	Permissions *UserPermissionModel `gorm:"foreignKey:UserId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"permissions"`
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

// MarshalJSON leaves the password hash out of every response that carries a user.
// It went out on login, on register, and on anything else returning a user,
// because UserContent's Password is tagged "password" - and that same tag is how
// login and register read the password out of the request body, so the tag has to
// stay and the hash is dropped on the way out instead. The outer Password field is
// shallower than the embedded one, so it hides it, and omitempty leaves it out.
func (u User) MarshalJSON() ([]byte, error) {
	type user User // a defined type does not inherit this method, so no recursion
	return json.Marshal(struct {
		user
		Password string `json:"password,omitempty"`
	}{user: user(u)})
}

// MarshalJSON - the audit row carries the same hash; see User.MarshalJSON.
func (u UserAt) MarshalJSON() ([]byte, error) {
	type userAt UserAt
	return json.Marshal(struct {
		userAt
		Password string `json:"password,omitempty"`
	}{userAt: userAt(u)})
}
