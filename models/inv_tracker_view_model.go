package models

type InvTrackerView struct {
	Id          uint   `json:"id"`
	ItemCode    string `json:"item_code"`
	GeneralName string `json:"general_name"`
	Brand       string `json:"brand"`
	ItemDesc    string `json:"item_desc"`
	Location    string `json:"location"`
	Qty         uint   `json:"qty"`
	Uom         string `json:"uom"`
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
