package models

type UserListView struct {
	UserName string `json:"user_name"`
}

func (UserListView) TableName() string {
	return "vw_get_all_user_list"
}
