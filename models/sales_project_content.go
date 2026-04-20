package models

type SalesProjectContentContent struct {
	// SHOULD BE THE TAB # / SET #
	BasedId              uint   `json:"based_id"`
	ItemDesignation      string `json:"item_designation"`
	Application          string `json:"application"`
	Additional           string `json:"additional"`
	Flow                 string `json:"flow"`
	Head                 string `json:"head"`
	Voltage              string `json:"voltage"`
	RPM                  string `json:"rpm"`
	HP                   string `json:"hp"`
	Phase                string `json:"phase"`
	NoOfSets             string `json:"no_of_sets"`
	NoOfPumpSet          string `json:"no_of_pump_set"`
	ItemSetDescription   string `json:"item_set_description"`
	ItemSetNotes         string `json:"item_set_notes"`
	AssignEngineerUserId uint   `json:"assign_engineer_user_id"`
	TemplateProjectId    uint   `json:"template_project_id"`
	IsWiring             *bool  `json:"is_wiring"`
}

type SalesProjectContent struct {
	ContentID uint `json:"content_id" gorm:"primaryKey"`
	SalesProjectContentContent
	SalesProjectContentFinal []SalesProjectContentFinal `json:"sales_project_content_final" gorm:"foreignKey:SalesProjectContentID;references:ContentID"`
}

func (SalesProjectContent) TableName() string {
	return "tbl_trans_sales_project_content"
}

type SalesProjectContentAt struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	RefID uint `json:"ref_id"`
	SalesProjectContentContent
	At
}

func (SalesProjectContentAt) TableName() string {
	return "z_tbl_trans_sales_project_content_at"
}
