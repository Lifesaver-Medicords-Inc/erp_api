package models

// Size Up - spec 5.1.4: "a manual list of candidate pumps (more than five allowed,
// scrollable)", and Final Selection is "limited to what is listed in Size Up". It hangs
// off the same content row as SalesProjectContentFinal (one list per Item/Set tab).
//
// This had no table, no model and no persistence at all before: the client built the grid
// in-session and GetSizeUpData() was dead code with no callers, so a saved quote always
// reopened with an empty Size Up - which in turn left Final Selection unconstrained.
type SalesProjectSizeUpContent struct {
	SalesProjectContentID uint   `json:"sales_project_content_id"`
	ItemID                uint   `json:"item_id"`
	Model                 string `json:"model"`
}

type SalesProjectSizeUp struct {
	ID uint `json:"id" gorm:"primaryKey"`
	SalesProjectSizeUpContent
}

func (SalesProjectSizeUp) TableName() string {
	return "tbl_trans_sales_project_size_up"
}

type SalesProjectSizeUpAt struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	RefID uint `json:"ref_id"`
	SalesProjectSizeUpContent
	At
}

func (SalesProjectSizeUpAt) TableName() string {
	return "z_tbl_trans_sales_project_size_up_at"
}
