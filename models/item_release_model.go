package models

type ItemReleaseContent struct {
	RequestDate    string `json:"request_date"`
	RequiredDate   string `json:"required_date"`
	ReleasedDate   string `json:"released_date"`
	ReferenceDocNo string `gorm:"size:50" json:"reference_doc_no"`
	RequestedBy    string `json:"requested_by"`
	ReceivedBy     string `json:"received_by"`
	ApprovedBy     string `json:"approved_by"`
	IssuedBy       string `json:"issued_by"`
	IsForward      *bool  `json:"is_forward"`
}

type ItemRelease struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"size:50" json:"doc_no"`
	ItemReleaseContent
	ItemReleaseDetails []ItemReleaseDetails `gorm:"foreignKey:ItemReleaseID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item_release_details,omitempty"`
}

func (ItemRelease) TableName() string {
	return "tbl_inv_item_release"
}

type ItemReleaseAt struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	RefId uint   `json:"ref_id"`
	DocNo string `gorm:"size:50;" json:"doc_no"`
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
	ItemDescription     string `json:"item_description"`
	RequiredQty         uint   `json:"required_qty"`
	RequiredUomID       string `json:"required_uom"`
	ReleasedQty         uint   `json:"released_qty"`
	ReleasedUomID       string `json:"released_uom"`
	SerialNo            string `json:"serial_no"`
	DeliveryPreference  string `json:"delivery_preference"`
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
