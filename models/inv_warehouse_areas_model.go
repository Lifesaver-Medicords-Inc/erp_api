package models

type WarehouseAreaContent struct {
	WarehouseNameId uint   `json:"warehouse_name_id"` // parent id
	UseType         string `json:"use_type"`
	Zone            string `json:"zone"`
	Area            string `json:"area"`
	Rack            string `json:"rack"`
	Level           string `json:"level"`
	Bins            string `json:"bins"`
	LocationCode    string `json:"location_code"`
	Notes           string `json:"notes"`

	// The vehicle this zone IS, when it is one. Spec 4.4.4: "Every vehicle also
	// exists as an OUTBOUND zone in the warehouse it is homed to", and 4.4.2: "if
	// the zone is a vehicle, the location code is the zone name only".
	//
	// Needed as a real link rather than inferred from UseType, because OUTBOUND
	// covers "the mezzanine and every vehicle" (4.4.2) - so use type alone cannot
	// tell a truck from a loading bay. That distinction is not cosmetic: Item
	// Release is the only document allowed to drive stock negative and only
	// against a VEHICLE zone (10.5), so anything deciding that has to be able to
	// answer "is this bin a vehicle?" exactly. Matching on zone names would make
	// a renamed zone silently change what the warehouse is permitted to do.
	//
	// Zero on every ordinary zone.
	VehicleId uint `json:"vehicle_id"`
}

type WarehouseArea struct {
	ID uint `gorm:"primarykey" json:"id"`
	WarehouseAreaContent
}

func (WarehouseArea) TableName() string {
	return "tbl_inv_warehouse_area"
}

type WarehouseAreaAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	WarehouseAreaContent
	At
}

func (WarehouseAreaAt) TableName() string {
	return "z_tbl_inv_warehouse_area_at"
}

type WarehouseAreaView struct {
	WarehouseNameId uint   `json:"warehouse_name_id"` // parent id
	Zone            string `json:"zone"`
	Area            string `json:"area"`
	Rack            string `json:"rack"`
	Level           string `json:"level"`
	Bins            string `json:"bins"`
	LocationCode    string `json:"location_code"`
	WarehouseName   string `json:"warehouse_name"`
}

func (WarehouseAreaView) TableName() string {
	return "vw_get_bin_loc_pick_activity"
}
