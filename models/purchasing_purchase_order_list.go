package models

type PurchaseOrderList struct {
	ID             int    `json:"id"`
	DocNo          string `json:"doc_no"`
	SupplierName   string `json:"supplier_name"`
	TotalAmountDue string `json:"total_amount_due"`
	LeadTime       string `json:"lead_time"`
}

type PurchasingActivePO struct {
	PurchaseOrderList
}

func (PurchasingActivePO) TableName() string {
	return "vw_get_purchasing_active_po"
}

type PurchasingClosedPO struct {
	PurchaseOrderList
	ReceivingReportID string `json:"receiving_report_id"`
	ReceivingReportNo string `json:"receiving_report_no"`
}

func (PurchasingClosedPO) TableName() string {
	return "vw_get_purchasing_closed_po"
}
