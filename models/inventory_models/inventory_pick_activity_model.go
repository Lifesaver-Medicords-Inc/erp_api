package inventory_models

import "github.com/pierceperado/smpc/models"

type PickActivityContent struct {
	SalesOrderId int    `json:"sales_order_id"`
	Customer     string `json:"customer"`
	CustomerCode string `json:"customer_code"`
	ReferenceSo  string `json:"reference_so"`
	SalesPerson  string `json:"sales_person"`
	PreparedBy   string `json:"prepared_by"`
	PickedBy     string `json:"picked_by"`
}

type PickActivity struct {
	ID    uint `gorm:"primarykey" json:"id"`
	DocNo int  `json:"doc_no"`
	PickActivityContent
}

func (PickActivity) TableName() string {
	return "tbl_inv_pick_activity2"
}

type PickActivityAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PickActivityContent
	models.At
}

func (PickActivityAt) TableName() string {
	return "z_tbl_inv_pick_activity_at2"
}

type PickActivityDetailsContent struct {
	PickActivityId      uint   `json:"pick_activity_id"`
	SalesOrderDetailsId uint   `json:"sales_order_details_id"`
	ItemId              uint   `json:"item_id"`
	ItemCode            string `json:"item_code"`
	ItemDescription     string `json:"item_description"`
	PickQty             uint   `json:"pick_qty"`
	PickUom             string `json:"pick_uom"`
	ActualQty           int    `json:"actual_qty"`
	ActualUom           string `json:"actual_uom"`
	BinLocation         string `json:"bin_location"`
	Warehouse           string `json:"warehouse"`
	WarehouseId         uint   `json:"warehouse_id"`
	HasActual           *bool  `json:"has_actual"`
}

type PickActivityDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	PickActivityDetailsContent
}

func (PickActivityDetails) TableName() string {
	return "tbl_inv_pick_activity_details2"
}

type PickActivityDetailsGet struct {
	PickActivityDetails
	LeftQty int    `json:"left_qty"`
	LeftUom string `json:"left_uom"`
}

func (PickActivityDetailsGet) TableName() string {
	return "vw_get_pick_activity_details"
}

type PickActivityDetailsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PickActivityDetailsContent
	models.At
}

func (PickActivityDetailsAt) TableName() string {
	return "z_tbl_inv_pick_activity_details_at2"
}

type PickActivityLocationsContent struct {
	PickActivityId        uint `json:"pick_activity_id"`
	PickActivityDetailsId uint `json:"pick_activity_details_id"`
	BinId                 uint `json:"bin_id"`
	SelectedQty           int  `json:"selected_qty"`
}

type PickActivityLocations struct {
	ID uint `gorm:"primarykey" json:"id"`
	PickActivityLocationsContent
}

type PickActivityLocationsAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	PickActivityLocationsContent
	models.At
}

func (PickActivityLocationsAt) TableName() string {
	return "z_tbl_inv_pick_activity_locations_at2"
}

func (PickActivityLocations) TableName() string {
	return "tbl_inv_pick_activity_locations2"
}

type PickActivityBody struct {
	PickActivity          PickActivity            `json:"pick_activity"`
	PickActivityDetails   []PickActivityDetails   `json:"pick_activity_details"`
	PickActivityLocations []PickActivityLocations `json:"pick_activity_locations"`
}

type PickActivityGet struct {
	PickActivity          []PickActivity           `json:"pick_activity"`
	PickActivityDetails   []PickActivityDetailsGet `json:"pick_activity_details"`
	PickActivityLocations []PickActivityLocations  `json:"pick_activity_locations"`
}
