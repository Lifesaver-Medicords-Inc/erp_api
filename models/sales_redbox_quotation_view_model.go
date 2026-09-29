package models

// SalesRedBoxQuotationView is one quote card of the Sales red box, first half (spec 9): the
// latest copy of a quote that has no SO yet, with who raised it and when (from its first copy).
type SalesRedBoxQuotationView struct {
	ID           uint   `json:"id"`
	DocumentNo   string `json:"document_no"`
	VersionNo    string `json:"version_no"`
	SubVersionNo string `json:"sub_version_no"`
	QuoteCode    string `json:"quote_code"`
	IsFinalized  bool   `json:"is_finalized"`
	ProjectName  string `json:"project_name"`
	CustomerID   uint   `json:"customer_id"`
	CreatedBy    string `json:"created_by"`
	Date         string `json:"date"`
}

func (SalesRedBoxQuotationView) TableName() string {
	return "vw_get_sales_redbox_quotations"
}
