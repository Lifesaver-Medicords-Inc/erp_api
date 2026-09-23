package models

type SalesProjectContentContent struct {
	// SHOULD BE THE TAB # / SET #
	BasedId         uint   `json:"based_id"`
	ItemDesignation string `json:"item_designation"`
	Application     string `json:"application"`
	Additional      string `json:"additional"`
	Flow            string `json:"flow"`
	Head            string `json:"head"`

	// The unit each of those two figures is quoted in, held as an ID rather than as its
	// text. The lists are hard-coded in the client for now - FLOW 1=gpm lpm, 2=m^2/hr,
	// 3=lps; HEAD 1=psi, 2=ft, 3=m (user decision 2026-09-23) - and the whole point of
	// storing the id is that turning either into a Setup list later is then a change to
	// the list, not a migration of every project quotation ever saved. The ids are fixed:
	// append, never renumber.
	//
	// Named after the controls that carry them, cmb_flow and cmb_head, because that is
	// what decides the key: Helpers.GetControlsValues maps a combo tagged DYNAMIC to
	// <its name without cmb_>_id. 0 means nothing was chosen and the client shows the
	// first entry.
	//
	// Before this the two combos were unnamed, unread and unsaved - whatever the user
	// picked was discarded.
	FlowId               uint   `json:"flow_id"`
	HeadId               uint   `json:"head_id"`
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
