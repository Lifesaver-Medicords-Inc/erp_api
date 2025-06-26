package accounting_models

type ChartOfAccountViewList struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ClassId   uint   `json:"class_id"`
	GroupId   uint   `json:"group_id"`
	ClassName string `json:"class_name"`
	GroupName string `json:"group_name"`
}

func (ChartOfAccountViewList) TableName() string {
	return "vw_get_chart_of_account"
}
