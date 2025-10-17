package models

type ReceivingReportContent2 struct {
	SupplierName    string `json:"supplier_name"`
	SupplierCode    string `json:"supplier_code"`
	DateReceived    string `json:"date_received"`
	Address         string `json:"address"`
	SupplierID      uint   `json:"supplier_id"`
	DOC             string `json:"doc"` //autogen
	RefDOC          string `json:"ref_doc"`
	PreparedBy      string `json:"prepared_by"`
	PurchaseOrderID uint   `json:"purchase_order_id"` //this is where the PO's will base
	WarehouseName   string `json:"warehouse_name"`
	WarehouseId     uint   `json:"warehouse_id"`
}

type ReceivingReport2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	// Code string `gorm:"unique" json:"code"`
	ReceivingReportContent2
}

func (ReceivingReport2) TableName() string {
	return "tbl_inv_warehouse_receiving_report2"
}

type ReceivingReportAt2 struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportContent2
	At
}

func (ReceivingReportAt2) TableName() string {
	return "z_tbl_inv_warehouse_receiving_report_at2"
}

type ReceivingReportDetailsContent2 struct {
	ReceivingReportId  uint   `json:"receiving_report_id"` //parent id
	PodId              uint   `json:"pod_id"`
	ItemID             uint   `json:"item_id"`
	ItemCode           string `json:"item_code"`
	ItemDescription    string `json:"item_description"`
	OrderedQty         string `json:"ordered_qty"`
	OrderedUom         string `json:"ordered_uom"`
	ReceivedQty        string `json:"received_qty"`
	ReceivedUom        string `json:"received_uom"`
	RejectedQty        string `json:"rejected_qty"`
	RejectedUom        string `json:"rejected_uom"`
	ReasonForRejection string `json:"reason_for_rejection"`
	SerialNumber       string `json:"serial_number"`
	BinLocation        string `json:"bin_location"`
	RefId              uint   `json:"ref_id"` //PO id
}

type ReceivingReportDetails2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	ReceivingReportDetailsContent2
}

func (ReceivingReportDetails2) TableName() string {
	return "tbl_inv_warehouse_receiving_report_details2"
}

type ReceivingReportDetailsAt2 struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportDetailsContent2
	At
}

func (ReceivingReportDetailsAt2) TableName() string {
	return "z_tbl_inv_warehouse_receiving_report_details_at2"
}

type ReceivingHistoryContent struct {
	PurchaseOrderID          uint   `json:"purchase_order_id"`
	PurchaseOrderDetailsID   uint   `json:"purchase_order_details_id"`
	ItemID                   uint   `json:"item_id"`
	ItemCode                 string `json:"item_code"`
	ReceivingReportID        uint   `json:"receiving_report_id"`
	ReceivingReportDetailsID uint   `json:"receiving_report_details_id"`
	OrderedQty               string `json:"ordered_qty"`
	ReceivedQty              string `json:"received_qty"`
	RejectedQty              string `json:"rejected_qty"`
	DateReceived             string `json:"date_received"`
	IsComplete               *bool  `json:"is_complete"`
}

type ReceivingHistory struct {
	ID uint `gorm:"primarykey" json:"id"`
	// Code string `gorm:"unique" json:"code"`
	ReceivingHistoryContent
}

func (ReceivingHistory) TableName() string {
	return "tbl_inv_warehouse_receiving_history"
}

type ReceivingHistoryAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingHistoryContent
	At
}

func (ReceivingHistoryAt) TableName() string {
	return "z_tbl_inv_warehouse_receiving_history_at"
}

type PurchaseOrderView struct {
	Id           uint   `json:"id"`
	SupplierId   uint   `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`
	SupplierCode string `json:"supplier_code"`
	DocNo        string `json:"doc_no"`
}

func (PurchaseOrderView) TableName() string {
	return "vw_get_purchase_order"
}
