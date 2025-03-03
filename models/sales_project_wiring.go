package models

type SalesProjectWiringContent struct {
	// Parent ID references to Sales Quotation Model
	BasedId     uint   `json:"based_id"`
	Materials   string `json:"materials"`
	AmpReq      string `json:"amp_req"`
	Description string `json:"description"`
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
	SalesProjectWiring
	At
}

func (SalesProjectWiringAt) TableName() string {
	return "z_tbl_trans_sales_project_wiring_at"
}
