package models

type EngineeringRedboxQuotationListContent struct {
	Id             int    `json:"id"`
	ClientName     string `json:"client_name"`
	SalesQuotation string `json:"sales_quotation"`
	Status         string `json:"status"`
	ProjectName    string `json:"project_name"`
	SalesExecutive string `json:"sales_executive"`
	Remark         string `json:"remark"`
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
