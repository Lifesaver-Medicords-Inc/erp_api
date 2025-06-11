package models

type ReceivingReportContent struct {
	SupplierName    string `json:"supplier_name"`
	Code            string `json:"code"`
	DateReceived    string `json:"date_received"`
	Address         string `json:"address"`
	SupplierID      uint   `json:"supplier_id"`
	DOC             string `json:"doc"` //autogen
	RefDOC          string `json:"ref_doc"`
	PreparedBy      string `json:"prepared_by"`
	PurchaseOrderID uint   `json:"purchase_order_id"` //this is where the PO's will base
}

type ReceivingReport struct {
	ID uint `gorm:"primarykey" json:"id"`
	// Code string `gorm:"unique" json:"code"`
	ReceivingReportContent
}

func (ReceivingReport) TableName() string {
	return "tbl_receiving_report"
}

type ReceivingReportAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportContent
	At
}

func (ReceivingReportAt) TableName() string {
	return "z_tbl_receiving_report_at"
}
