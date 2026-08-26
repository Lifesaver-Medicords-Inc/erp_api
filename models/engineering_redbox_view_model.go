package models

type EngineeringRedboxQuotationListContent struct {
	Id             int    `json:"id"`
	ClientName     string `json:"client_name"`
	SalesQuotation string `json:"sales_quotation"`
	Status         string `json:"status"`
	ProjectName    string `json:"project_name"`
	SalesExecutive string `json:"sales_executive"`
	Remark         string `json:"remark"`
	// Added for Phase 4 item 4.1 - lets a future per-engineer scoped query (the
	// Sales Quotation List itself, or a scoped WS route mirroring Job Order's own
	// /job_order/:userId) filter with conditions{"requested_engr_id": ...} via the
	// existing GetSortedEngineeringRedboxQuotationList(conditions) plumbing,
	// without needing a new stored procedure.
	RequestedEngrId uint `json:"requested_engr_id"`
}

type EngineeringRedboxQuotationListView struct {
	EngineeringRedboxQuotationListContent
}

func (EngineeringRedboxQuotationListView) TableName() string {
	return "vw_get_engineering_redbox_quotation_list"
}

type EngineeringRedboxJobOrderContent struct {
	Id          int    `json:"id"`
	ClientName  string `json:"client_name"`
	DocumentNo  string `json:"document_no"`
	Items       int    `json:"items"`
	ProjectName string `json:"project_name"`
	DueDate     string `json:"due_date"`
	Type        string `json:"type"`
}

type EngineeringRedboxJobOrderView struct {
	EngineeringRedboxJobOrderContent
}

func (EngineeringRedboxJobOrderView) TableName() string {
	return "vw_get_engineering_redbox_job_order"
}
