package inventory_models

type UserListView struct {
	UserId   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

func (UserListView) TableName() string {
	return "vw_get_user_item_req"
}
