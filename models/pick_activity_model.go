package models

type PickActivityContent struct {
	Customer        string `json:"customer"`
	Code            string `json:"code"`
	ReferenceSo     string `json:"reference_so"`
	SalesPerson     string `json:"sales_person"`
	PreparedBy      string `json:"prepared_by"`
	PickedBy        string `json:"picked_by"`
	TransactionDate string `json:"transaction_date"`
}
type PickActivity struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	DocNo string `gorm:"unique" json:"doc_no"`
	PickActivityContent
}

func (PickActivity) TableName() string {
	return "tbl_trans_pick_activity"
}

type PickActivityAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PickActivityContent
	At
}

func (PickActivityAt) TableName() string {
	return "z_tbl_trans_pick_activity_at"
}

type PickActivityDetailsContent struct {
	ItemId          uint   `json:"item_id"`
	ItemCode        string `json:"item_code"`
	ItemDescription string `json:"item_description"`
	SOId            uint   `json:"so_id"`
	SODId           uint   `json:"sod_id"`
	LeftQty         uint   `json:"left_qty"`
	LeftUom         string `json:"left_uom"`
	PickQty         uint   `json:"pick_qty"`
	PickUom         string `json:"pick_uom"`
	ActualQty       uint   `json:"actual_qty"`
	ActualUom       string `json:"actual_uom"`
	BinLocation     string `json:"bin_location"`
	WarehouseId     uint   `json:"warehouse_id"`
}
type PickActivityDetails struct {
	ID   uint `gorm:"primarykey" json:"id"`
	PaId uint `json:"pa_id"`
	PickActivityDetailsContent
}

func (PickActivityDetails) TableName() string {
	return "tbl_trans_pick_activity_details"
}

type PickActivityDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PickActivityDetailsContent
	At
}

func (PickActivityDetailsAt) TableName() string {
	return "z_tbl_trans_pick_activity_details_at"
}

type PickActivityLocationContent struct {
	PaId        uint   `json:"pa_id"`
	PaDetailsId uint   `json:"pa_details_id"`
	StockQty    uint   `json:"stock_qty"`
	ActualQty   uint   `json:"actual_qty"`
	ActualUom   string `json:"actual_uom"`
	Location    string `json:"location"`
	WarehouseId uint   `json:"warehouse_id"`
	ItemId      uint   `json:"item_id"`
}

type PickActivityLocation struct {
	ID uint `gorm:"primarykey" json:"id"`
	PickActivityLocationContent
}

func (PickActivityLocation) TableName() string {
	return "tbl_trans_pick_activity_location"
}

type PickActivityLocationAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PickActivityLocationContent
	At
}

func (PickActivityLocationAt) TableName() string {
	return "z_tbl_trans_pick_activity_location_at"
}

type PickActivityHistoryContent struct {
	SOId            uint   `json:"so_id"`
	SODId           uint   `json:"sod_id"`
	RefDoc          string `json:"ref_doc"`
	PAId            uint   `json:"pa_id"`
	PADId           uint   `json:"pad_id"`
	ItemID          uint   `json:"item_id"`
	LeftQty         uint   `json:"left_qty"`
	PickQty         uint   `json:"pick_qty"`
	TransactionDate string `json:"transaction_date"`
}
type PickActivityHistory struct {
	ID uint `gorm:"primarykey" json:"id"`
	PickActivityHistoryContent
}

func (PickActivityHistory) TableName() string {
	return "tbl_trans_pick_activity_history"
}

type PickActivityHistoryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PickActivityHistoryContent
	At
}

func (PickActivityHistoryAt) TableName() string {
	return "z_tbl_trans_pick_activity_history_at"
}
