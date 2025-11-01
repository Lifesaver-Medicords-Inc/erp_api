package models

type ItemRequestContent struct {
	ReqDept      string `json:"req_dept"`
	Purpose      string `json:"purpose"`
	ReqDate      string `json:"req_date"`
	RequiredDate string `json:"required_date"`
	IssueDate    string `json:"issue_date"`
	RefDoc       string `json:"ref_doc"`
	ReqBy        string `json:"req_by"`
	ReceivedBy   string `json:"received_by"`
	ApprovedBy   string `json:"approved_by"`
	IssuedBy     string `json:"issued_by"`
	IsForward    *bool  `json:"is_forward"`
}
type ItemRequest struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	DocNo string `gorm:"unique" json:"doc_no"`
	ItemRequestContent
}

func (ItemRequest) TableName() string {
	return "tbl_item_request"
}

type ItemRequestAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemRequestContent
	At
}

func (ItemRequestAt) TableName() string {
	return "z_tbl_item_request_at"
}

type ItemRequestDetailsContent struct {
	ItemId          uint   `json:"item_id"`
	ItemDescription string `json:"item_description"`
	ReqQty          uint   `json:"req_qty"`
	ReqUom          string `json:"req_uom"`
	IssuedQty       uint   `json:"issued_qty"`
	TotalReq        uint   `json:"total_req"`
	TotalIssued     uint   `json:"total_issued"`
	IssuedUom       string `json:"issued_uom"`
	SerialNo        string `json:"serial_no"`
	Remarks         string `json:"remarks"`
}
type ItemRequestDetails struct {
	ID   uint `gorm:"primarykey" json:"id"`
	IrId uint `json:"ir_id"`
	ItemRequestDetailsContent
}

func (ItemRequestDetails) TableName() string {
	return "tbl_item_request_details"
}

type ItemRequestDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemRequestDetailsContent
	At
}

func (ItemRequestDetailsAt) TableName() string {
	return "z_tbl_item_request_details_at"
}

type ItemRequestLocationContent struct {
	IrId        uint   `json:"ir_id"`
	IrDetailsId uint   `json:"ir_details_id"`
	IssuedQty   uint   `json:"issued_qty"`
	IssuedUom   string `json:"issued_uom"`
	Location    string `json:"location"`
	WarehouseId uint   `json:"warehouse_id"`
}

type ItemRequestLocation struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemRequestLocationContent
}

func (ItemRequestLocation) TableName() string {
	return "tbl_item_request_location"
}

type ItemRequestLocationAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemRequestLocationContent
	At
}

func (ItemRequestLocationAt) TableName() string {
	return "z_tbl_item_request_location_at"
}
