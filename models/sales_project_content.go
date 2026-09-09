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

	// §5.1.4: "right-click a tab to exclude it -> the tab highlights light red and its
	// items are excluded from gross-sales computation AND from the printed proposal"
	// (negative test 35). The flag lived only in the client, as a HashSet<TabPage> held in
	// memory, so it was lost the moment the quotation was reloaded and was invisible to the
	// print modal - which fetches its own data and never sees another form's controls.
	//
	// Stored on the content row because that IS the tab: one row per Item/Set. It travels
	// in the payload the quote already sends, so persisting it costs no extra API call and
	// no extra query - one column on a row that was being written anyway.
	//
	// Pointer, matching IsWiring: a body that omits the field leaves the stored value alone
	// rather than silently un-excluding a tab.
	IsExcluded *bool `json:"is_excluded"`
}

type SalesProjectContent struct {
	ContentID uint `json:"content_id" gorm:"primaryKey"`
	SalesProjectContentContent
	SalesProjectContentFinal []SalesProjectContentFinal `json:"sales_project_content_final" gorm:"foreignKey:SalesProjectContentID;references:ContentID"`
	SalesProjectSizeUp       []SalesProjectSizeUp       `json:"sales_project_size_up" gorm:"foreignKey:SalesProjectContentID;references:ContentID"`
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
