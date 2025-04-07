package models

type CRMView struct {
	ID         uint   `json:"id"`
	BranchName string `json:"branch_name"`
	Number     string `json:"number"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Tag        string `json:"tag"`
	Date       string `json:"date"`
	Remark     string `json:"remark"`
	CRM_ID     uint   `json:"crm_id"`
}

// test
func (CRMView) TableName() string {
	return "vw_get_CRM"
}
