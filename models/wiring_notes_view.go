package models

type WiringNotes struct {
	WiringID             uint   `json:"wiring_id"`
	BasedID              uint   `json:"based_id"`
	Materials            string `json:"materials"`
	ItemDescription      string `json:"item_description"`
	NumOfWiresSet        string `json:"num_of_wires_set"`
	NumOfQtySet          string `json:"num_of_qty_set"`
	DistanceTravelledSet string `json:"distance_travelled_set"`
	AllowanceWireSet     string `json:"allowance_wire_set"`
	NumOfSets            string `json:"num_of_sets"`
	TotalQty             uint   `json:"total_qty"`
	WiringNote           string `json:"wiring_note"`
	NoteID               uint   `json:"note_id"`
}

func (WiringNotes) TableName() string {
	return "GetWiringNotes"
}
