package models

type SalesProjectWiringContent struct {
	// Parent ID references to Sales Quotation Model
	BasedId              uint    `json:"based_id"`
	Materials            string  `json:"materials"`
	AmpReq               string  `json:"amp_req"`
	WireReq              string  `json:"wire_req"`
	Description          string  `json:"description"`
	NumOfWiresSet        string  `json:"num_of_wires_set"`
	NumOfQtySet          string  `json:"num_of_qty_set"`
	DistanceTravelledSet string  `json:"distance_travelled_set"`
	AllowanceWireSet     string  `json:"allowance_wire_set"`
	Qty                  uint    `json:"qty"`
	NumOfSets            string  `json:"num_of_sets"`
	TotalQty             uint    `json:"total_qty"`
	Cost                 float64 `json:"cost"`
	TotalCost            float64 `json:"total_cost"`
}

type SalesProjectWiring struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesProjectWiringContent
}

func (SalesProjectWiring) TableName() string {
	return "tbl_trans_sales_project_wiring"
}

type SalesProjectWiringAt struct {
	WiringID uint `gorm:"primarykey" json:"wiring_id"`
	RefId    uint `json:"ref_id"`
	SalesProjectWiringContent
	At
}

func (SalesProjectWiringAt) TableName() string {
	return "z_tbl_trans_sales_project_wiring_at"
}
