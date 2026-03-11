package dispatching_models

type IRDetailsApprovedSOView struct {
	ItemReleaseID uint   `json:"item_release_id"`
	ItemID        uint   `json:"item_id"`
	ReleasedQty   int    `json:"released_qty"`
	ReleasedUom   string `json:"released_uom"`
	ItemCode      string `json:"item_code"`
	SerialNo      string `json:"serial_no"`
}
