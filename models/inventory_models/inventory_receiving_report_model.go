package inventory_models

import "github.com/pierceperado/smpc/models"

type ReceivingReportContent struct {
	Supplier         string `json:"supplier"`
	SupplierCode     string `json:"supplier_code"`
	SupplierID       uint   `json:"supplier_id"`
	DateReceived     string `json:"date_received"`
	RefDoc           string `json:"ref_doc"`
	PreparedBy       string `json:"prepared_by"`
	PurchaseOrderID  uint   `json:"purchase_order_id"`
	Warehouse        string `json:"warehouse"`
	WarehouseAddress string `json:"warehouse_address"`
	WarehouseId      uint   `json:"warehouse_id"`
}

type ReceivingReport struct {
	ID    uint `gorm:"primarykey" json:"id"`
	DocNo int  `json:"doc_no"`
	ReceivingReportContent
}

func (ReceivingReport) TableName() string {
	return "tbl_inv_receiving_report"
}

type ReceivingReportAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportContent
	models.At
}

func (ReceivingReportAt) TableName() string {
	return "z_tbl_inv_receiving_report_at"
}

type ReceivingReportDetailsContent struct {
	ReceivingReportId      uint   `json:"receiving_report_id"`
	PurchaseOrderDetailsId uint   `json:"purchase_order_details_id"`
	ItemID                 uint   `json:"item_id"`
	ItemCode               string `json:"item_code"`
	ItemDesc               string `json:"item_desc"`
	OrderedQty             int    `json:"ordered_qty"`
	OrderedUom             string `json:"ordered_uom"`
	ReceivedQty            int    `json:"received_qty"`
	ReceivedUom            string `json:"received_uom"`
	SerialNumber           string `json:"serial_number"`
	WarehouseId            uint   `json:"warehouse_id"`
	BinLocation            string `json:"bin_location"`
	RejectedQty            int    `json:"rejected_qty"`
	RejectedUom            string `json:"rejected_uom"`
	ReasonForRejection     string `json:"reason_for_rejection"`
}

type ReceivingReportDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	ReceivingReportDetailsContent
}

func (ReceivingReportDetails) TableName() string {
	return "tbl_inv_receiving_report_details"
}

type ReceivingReportDetailsGet struct {
	ReceivingReportDetails
	RemainingQty int    `json:"remaining_qty"`
	RemainingUom string `json:"remaining_uom"`
}

func (ReceivingReportDetailsGet) TableName() string {
	return "vw_get_receiving_report_details"
}

type ReceivingReportDetailsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportDetailsContent
	models.At
}

func (ReceivingReportDetailsAt) TableName() string {
	return "z_tbl_inv_receiving_report_details_at"
}

type ReceivingReportBody struct {
	ReceivingReport        ReceivingReport          `json:"receiving_report"`
	ReceivingReportDetails []ReceivingReportDetails `json:"receiving_report_details"`
}

type ReceivingReportGet struct {
	ReceivingReport        []ReceivingReport           `json:"receiving_report"`
	ReceivingReportDetails []ReceivingReportDetailsGet `json:"receiving_report_details"`
}
