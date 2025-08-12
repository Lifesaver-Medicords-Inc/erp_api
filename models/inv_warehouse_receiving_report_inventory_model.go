package models

type ReceivingReportInventoryContent struct {
	ReceivingReportId uint   `json:"receiving_report_id"` //parent id
	ItemCode          string `json:"item_code"`
	ItemDescription   string `json:"item_description"`
	OrderedQty        string `json:"ordered_qty"`
	OrderedUom        string `json:"ordered_uom"`
	SerialNumber      string `json:"serial_number"`
	BinLocation       string `json:"bin_location"`
	RefId             uint   `json:"ref_id"` //PO id
}

type ReceivingReportInventory struct {
	ID uint `gorm:"primarykey" json:"id"`
	ReceivingReportInventoryContent
}

func (ReceivingReportInventory) TableName() string {
	return "tbl_inv_warehouse_receiving_report_inventory"
}

type ReceivingReportInventoryAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ReceivingReportInventoryContent
	At
}

func (ReceivingReportInventoryAt) TableName() string {
	return "z_tbl_inv_warehouse_receiving_report_inventory"
}
