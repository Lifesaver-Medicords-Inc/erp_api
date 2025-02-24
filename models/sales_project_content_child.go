package models

type SalesProjectContentChildContent struct {
	// SHOULD BE THE TAB # / SET #
	BasedID     uint   `json:"based_id"`
	Flow        string `json:"flow"`
	Head        string `json:"head"`
	Voltage     string `json:"voltage"`
	RPM         string `json:"rpm"`
	HP          string `json:"hp"`
	Phase       string `json:"phase"`
	NoOfSets    uint   `json:"no_of_sets"`
	NoOfPumpSet uint   `json:"no_of_pump_set"`
}

type SalesProjectContentChild struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesProjectContentChildContent
}

func (SalesProjectContentChild) TableName() string {
	return "tbl_trans_sales_project_content_child"
}

type SalesProjectContentChildAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesProjectContentChild
	At
}

func (SalesProjectContentChildAt) TableName() string {
	return "z_tbl_trans_sales_project_content_child_at"
}
