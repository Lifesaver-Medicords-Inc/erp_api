package models

type ItemReleaseContent struct {
	RequestDate    string `json:"request_date"`
	RequiredDate   string `json:"required_date"`
	ReleasedDate   string `json:"released_date"`
	SalesOrderId   uint   `json:"sales_order_id"`
	ReferenceDocNo string `gorm:"size:50" json:"reference_doc_no"`
	RequestedBy    string `json:"requested_by"`
	ReceivedBy     string `json:"received_by"`
	ApprovedBy     string `json:"approved_by"`
	IssuedBy       string `json:"issued_by"`
	IsForward      *bool  `json:"is_forward"`
}

type ItemRelease struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"-:migration" json:"doc_no"`
	ItemReleaseContent
	ItemReleaseDetails *[]ItemReleaseDetails `gorm:"foreignKey:ItemReleaseID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item_release_details,omitempty"`
}

func (ItemRelease) TableName() string {
	return "tbl_inv_item_release"
}

type ItemReleaseAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DocNo int  `gorm:"-:migration" json:"doc_no"`
	ItemReleaseContent
	At
}

func (ItemReleaseAt) TableName() string {
	return "z_tbl_inv_item_release_at"
}

type ItemReleaseDetailsContent struct {
	ItemReleaseID       uint   `gorm:"not null;index" json:"item_release_id"`
	SalesOrderID        uint   `json:"sales_order_id"`
	SalesOrderDetailsID uint   `json:"sales_order_details_id"`
	ItemID              uint   `json:"item_id"`
	ItemCode            string `json:"item_code"`
	ItemDescription     string `json:"item_description"`
	RequiredQty         uint   `json:"required_qty"`
	RequiredUomID       string `json:"required_uom"`
	ReleasedQty         uint   `json:"released_qty"`
	ReleasedUomID       string `json:"released_uom"`
	SerialNo            string `json:"serial_no"`
	DeliveryPreference  string `json:"delivery_preference"`

	// Per-bin breakdown of ReleasedQty, from the dispatching app's PickActivity picker.
	// JSON-only (gorm:"-") - never touched by GORM's association save, so a plain
	// detail-row upsert can never silently re-deduct stock behind ApplyItemReleaseLocations'
	// back. See item_release_service.go's ApplyItemReleaseLocations/
	// RestoreItemReleaseLocations for the code that actually persists these rows and
	// moves stock - this field only carries them from the request body to that code.
	Locations []ItemReleaseLocations `gorm:"-" json:"locations,omitempty"`
}

type ItemReleaseDetails struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ItemReleaseDetailsContent
}

func (ItemReleaseDetails) TableName() string {
	return "tbl_inv_item_release_details"
}

type ItemReleaseDetailsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemReleaseDetailsContent
	At
}

func (ItemReleaseDetailsAt) TableName() string {
	return "z_tbl_inv_item_release_details_at"
}

// ItemReleaseLocations is the bin-level record behind one ItemReleaseDetails line -
// same shape/role as ItemRequestLocations and PickActivityLocations (BinId + qty taken
// from that specific tbl_inv_item_stocks row), added so Item Release can finally call
// DeductStockWithTx at all (previously: never did, for any bin - see the service file).
//
// Deliberately keyed by this row's own ID (not the parent ItemRelease's ID) when it
// talks to DeductStockWithTx/RestoreStockWithTx below, unlike Item Request/Pick Activity
// which both key on their *header* ID - that shared-ID choice makes ReleaseLotsFIFO
// (which reverses everything recorded under one refType+refId, with no finer
// granularity) reverse an entire document's FIFO cost lots even when only one line/bin
// is being restored. Using this row's own ID keeps every deduct/restore scoped to
// exactly the bin it concerns. Not fixed on the other two here - out of scope for this
// change - just not repeated.
type ItemReleaseLocationsContent struct {
	ItemReleaseDetailsID uint `json:"item_release_details_id"`
	BinId                uint `json:"bin_id"`
	SelectedQty          int  `json:"selected_qty"`
}

type ItemReleaseLocations struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ItemReleaseLocationsContent
}

func (ItemReleaseLocations) TableName() string {
	return "tbl_inv_item_release_locations"
}

type ItemReleaseLocationsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemReleaseLocationsContent
	At
}

func (ItemReleaseLocationsAt) TableName() string {
	return "z_tbl_inv_item_release_locations_at"
}
