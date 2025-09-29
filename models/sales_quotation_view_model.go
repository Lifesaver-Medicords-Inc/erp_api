package models

type LatestQuotationView struct {
	SalesQuotation
}

func (LatestQuotationView) TableName() string {
	return "vw_latest_sales_quotation"
}
