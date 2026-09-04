package models

type SalesProjectContentFinalContent struct {
	SalesProjectContentID uint `json:"sales_project_content_id"`
	// ItemID added 2026-09-03. Without it, a final row reloaded from the database
	// came back with no item id, which broke ItemSetUC.SetFinalPumpData's own
	// duplicate guard (it compares the grid's hidden final_item_id against the
	// pump being added). After any reload that guard could never match, so
	// re-picking a pump already on the list appended a second row instead of
	// being rejected - and one save then wrote both. That is how content 14 ended
	// up with each of its two pumps stored twice (all four INSERTs in a single
	// transaction, per z_tbl_trans_sales_project_content_final_at).
	//
	// Size Up has carried item_id from the start, which is exactly why its own
	// picker dedupe (AddSizeUpRow) kept working and size-up rows never duplicated.
	ItemID  uint    `json:"item_id"`
	Final   string  `json:"final"`
	Fla     float64 `json:"fla"`
	Voltage float64 `json:"voltage"`
}

type SalesProjectContentFinal struct {
	ID uint `json:"id" gorm:"primaryKey"`
	SalesProjectContentFinalContent
}

func (SalesProjectContentFinal) TableName() string {
	return "tbl_trans_sales_project_content_final"
}

type SalesProjectContentFinalAt struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	RefID uint `json:"ref_id"`
	SalesProjectContentFinalContent
	At
}

func (SalesProjectContentFinalAt) TableName() string {
	return "z_tbl_trans_sales_project_content_final_at"
}
