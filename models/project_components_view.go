package models

type ProjectComponent struct {
	QuotationID               uint    `json:"quotation_id"`
	ProjectName               string  `json:"project_name"`
	CustomerID                uint    `json:"customer_id"`
	SetID                     uint    `json:"set_id"`
	ItemSetBasedOnQuotationID uint    `json:"item_set_based_on_quotation_id"`
	ItemSetName               string  `json:"item_set_name"`
	BomID                     uint    `json:"bom_id"`
	ItemsID                   uint    `json:"items_id"`
	Model                     string  `json:"model"`
	Qty                       uint    `json:"qty"`
	BasedOnSetID              uint    `json:"based_on_set_id"`
	Components                string  `json:"components"`
	NodeID                    uint    `json:"node_id"`
	NodeName                  string  `json:"node_name"`
	NodeOrder                 uint    `json:"node_order"`
	NodeType                  string  `json:"node_type"`
	ParentNodeID              uint    `json:"parent_node_id"`
	ItemName                  string  `json:"item_name"`
	ShortDesc                 string  `json:"short_desc"`
	Size                      string  `json:"size"`
	ItemID                    uint    `json:"item_id"`
	ComponentTotal            float64 `json:"component_total"`
	UnitOfMeasure             string  `json:"unit_of_measure"`
	CustomerName              string  `json:"customer_name"`
	Materials                 string  `json:"materials"`
	AmpReq                    uint    `json:"amp_req"`
	WireReq                   uint    `json:"wire_req"`
	WiringDescription         string  `json:"wiring_description"`
	NumOfWiresSet             uint    `json:"num_of_wires_set"`
	NumOfQtySet               uint    `json:"num_of_qty_set"`
	DistanceTravelledSet      uint    `json:"distance_travelled_set"`
	AllowanceWireSet          uint    `json:"allowance_wire_set"`
	WiringQty                 uint    `json:"wiring_qty"`
	NumOfSets                 uint    `json:"num_of_sets"`
	TotalQty                  uint    `json:"total_qty"`
	Cost                      uint    `json:"cost"`
	TotalCost                 uint    `json:"total_cost"`
	BOQID                     uint    `json:"boq_id"`
	Remarks                   string  `json:"remarks"`
	Notes                     string  `json:"notes"`
}

func (ProjectComponent) TableName() string {
	return "GetProjectComponents"
}
