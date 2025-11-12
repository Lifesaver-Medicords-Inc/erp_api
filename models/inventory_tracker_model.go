package models

type InvTrackerContent struct {
	RrId    uint   `json:"rr_id"`
	RrdId   uint   `json:"rrd_id"`
	PodId   uint   `json:"pod_id"`
	Remarks string `json:"remarks"`
}

type InvTracker struct {
	ID uint `gorm:"primarykey" json:"id"`
	InvTrackerContent
}

func (InvTracker) TableName() string {
	return "tbl_inv_tracker"
}

type InvTrackerAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	InvTrackerContent
	At
}

func (InvTrackerAt) TableName() string {
	return "z_tbl_inv_tracker_at"
}

type InvTrackerView struct {
	Id            uint   `json:"id"`
	RrId          uint   `json:"rr_id"`
	PodId         uint   `json:"pod_id"`
	ItemId        uint   `json:"item_id"`
	ItemCode      string `json:"item_code"`
	GeneralName   string `json:"general_name"`
	Brand         string `json:"brand"`
	ItemDesc      string `json:"item_desc"`
	Location      string `json:"location"`
	Qty           string `json:"qty"`
	Uom           string `json:"uom"`
	WarehouseName string `json:"warehouse_name"`
	Remarks       string `json:"remarks"`
	RemId         uint   `json:"rem_id"`
	WarehouseId   uint   `json:"warehouse_id"`
}

func (InvTrackerView) TableName() string {
	return "vw_get_inventory_tracker"
}

type WarehouseNameView struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

func (WarehouseNameView) TableName() string {
	return "tbl_inv_warehouse_name"
}

type InvLogbookView struct {
	Id              uint   `json:"id"`
	PodId           uint   `json:"pod_id"`
	ItemID          uint   `json:"item_id"`
	GeneralName     string `json:"general_name"`
	Brand           string `json:"brand"`
	ItemDescription string `json:"item_description"`
	ItemCategory    string `json:"item_category"`
	Calibration     string `json:"calibration"`
	ItemCode        string `json:"item_code"`
	ItemModel       string `json:"item_model"`
	Location        string `json:"location"`
	QtyIn           uint   `json:"qty_in"`
	QtyOut          uint   `json:"qty_out"`
	Uom             string `json:"uom"`
	Date            string `json:"date"`
	RrNo            string `json:"rr_no"`
	PoNo            string `json:"po_no"`
	SupplierName    string `json:"supplier_name"`
	Remarks         string `json:"remarks"`
}

func (InvLogbookView) TableName() string {
	return "vw_get_inventory_logbook"
}
