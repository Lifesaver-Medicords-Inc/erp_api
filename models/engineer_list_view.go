package models

type EngineerListView struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
}

func (EngineerListView) TableName() string {
	return "vw_get_users_engr_list"
}
