package models

type ReceivingReportDetailsContent struct {
	ReceivingReportId  uint   `json:"receiving_report_id"` //parent id
	ItemCode           string `json:"item_code"`
	ItemDescription    string `json:"item_description"`
	OrderedQty         string `json:"ordered_qty"`
	OrderedUom         string `json:"ordered_uom"`
	ReceivedQty        string `json:"received_qty"`
	ReceivedUom        string `json:"received_uom"`
	RejectedQty        string `json:"rejected_qty"`
	RejectedUom        string `json:"rejected_uom"`
	ReasonForRejection string `json:"reason_for_rejection"`
	RefId              uint   `json:"ref_id"` //PO id
}

type ReceivingReportDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	ReceivingReportDetailsContent
}

func (ReceivingReportDetails) TableName() string {
	return "tbl_inv_warehouse_receiving_report_details"
}

type ReceivingReportDetailsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportDetailsContent
	At
}

func (ReceivingReportDetailsAt) TableName() string {
	return "z_tbl_inv_warehouse_receiving_report_details_at"
}
