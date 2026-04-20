package inventory_models

import "github.com/pierceperado/smpc/models"

type ItemRequestContent struct {
	SalesOrderID   uint   `json:"sales_order_id"`
	RequestingDept string `json:"requesting_dept"`
	Purpose        string `json:"purpose"`
	RequestDate    string `json:"request_date"`
	RequiredDate   string `json:"required_date"`
	IssueDate      string `json:"issue_date"`
	RefDoc         string `json:"ref_doc"`
	RequestedBy    string `json:"requested_by"`
	ReceivedBy     string `json:"received_by"`
	ApprovedBy     string `json:"approved_by"`
	IssuedBy       string `json:"issued_by"`
	IsForward      *bool  `json:"is_forward"`
}

type ItemRequest struct {
	ID    uint `gorm:"primarykey" json:"id"`
	DocNo int  `json:"doc_no"`
	ItemRequestContent
}

func (ItemRequest) TableName() string {
	return "tbl_inv_item_request2"
}

type ItemRequestAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ItemRequestContent
	models.At
}

func (ItemRequestAt) TableName() string {
	return "z_tbl_inv_item_request_at2"
}

type ItemRequestDetailsContent struct {
	ItemRequestId       uint   `json:"item_request_id"`
	SalesOrderDetailsId uint   `json:"sales_order_details_id"`
	ItemId              uint   `json:"item_id"`
	ItemDescription     string `json:"item_description"`
	RequiredQty         int    `json:"required_qty"`
	RequiredUom         string `json:"required_uom"`
	IssuedQty           int    `json:"issued_qty"`
	IssuedUom           string `json:"issued_uom"`
	SerialNo            string `json:"serial_no"`
	Remarks             string `json:"remarks"`
	HasIssued           *bool  `json:"has_issued"`
}

type ItemRequestDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemRequestDetailsContent
}

func (ItemRequestDetails) TableName() string {
	return "tbl_inv_item_request_details2"
}

type ItemRequestDetailsGet struct {
	ItemRequestDetails
	RemainingQty int    `json:"remaining_qty"`
	RemainingUom string `json:"remaining_uom"`
}

func (ItemRequestDetailsGet) TableName() string {
	return "vw_get_item_request_details"
}

type ItemRequestDetailsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ItemRequestDetailsContent
	models.At
}

func (ItemRequestDetailsAt) TableName() string {
	return "z_tbl_inv_item_request_details_at2"
}

type ItemRequestLocationsContent struct {
	ItemRequestId        uint `json:"item_request_id"`
	ItemRequestDetailsId uint `json:"item_request_details_id"`
	BinId                uint `json:"bin_id"`
	SelectedQty          int  `json:"selected_qty"`
}

type ItemRequestLocations struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemRequestLocationsContent
}

type ItemRequestLocationsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ItemRequestLocationsContent
	models.At
}

func (ItemRequestLocationsAt) TableName() string {
	return "z_tbl_inv_item_request_locations_at2"
}

func (ItemRequestLocations) TableName() string {
	return "tbl_inv_item_request_locations2"
}

type ItemRequestBody struct {
	ItemRequest          ItemRequest            `json:"item_request"`
	ItemRequestDetails   []ItemRequestDetails   `json:"item_request_details"`
	ItemRequestLocations []ItemRequestLocations `json:"item_request_locations"`
}

type ItemRequestGet struct {
	ItemRequest          []ItemRequest           `json:"item_request"`
	ItemRequestDetails   []ItemRequestDetailsGet `json:"item_request_details"`
	ItemRequestLocations []ItemRequestLocations  `json:"item_request_locations"`
}
