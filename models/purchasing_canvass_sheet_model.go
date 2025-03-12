package models

type CanvassSheetContent struct {
	ItemName string `json:"item_name"`
	ItemCode string `json:"item_code"`
	DocNo    string `json:"doc_no"`
	Date     string `json:"date"`
	ReqQty   int    `json:"req_qty"`
	Stock    int    `json:"stock"`
}

type CanvassSheet struct {
	ID uint `gorm:"primarykey" json:"id"`
	CanvassSheetContent
}

func (CanvassSheet) TableName() string {
	return "tbl_purchasing_canvass_sheet"
}

type CanvassSheetAt struct {
	ID uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	CanvassSheetContent
	At
}

func (CanvassSheetAt) TableName() string {
	return "z_tbl_purchasing_canvass_sheet_at"
}

