package models

type InvTrackerContent struct {
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
	PodId         uint   `json:"pod_id"`
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
	return "vw_get_item_inventory"
}

type InvNameView struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

func (InvNameView) TableName() string {
	return "tbl_inv_warehouse_name"
}
