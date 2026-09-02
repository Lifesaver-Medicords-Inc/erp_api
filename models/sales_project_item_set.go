package models

type SalesProjectItemSetContent struct {
	BasedId   uint   `json:"based_id"`
	TabNumber string `json:"tab_number"`
}

type SalesProjectItemSet struct {
	// The actual DB column is "item_set_id" (confirmed in sql/views/GetProjectComponents.sql
	// and GetWiringNotes.sql, both of which reference tbl_trans_sales_project_item_set's
	// item_set_id column directly) - this was tagged "itemset_id" (no underscore), which
	// doesn't exist, so every insert into this table failed with "Invalid column name
	// 'itemset_id'".
	// autoIncrement is explicit: item_set_id IS an IDENTITY column in the DB, but with a
	// non-conventional PK name and an explicit column mapping GORM did not infer that,
	// so it included the column in the INSERT and wrote a literal 0. Every item set
	// therefore landed with item_set_id = 0, which cascaded: CreateProjectContent is
	// called with that id, so content.based_id became 0 too, and the client - which
	// loads a tab's content by "based_id = <itemset_id>" - could never match it back.
	ItemSetID uint `gorm:"primaryKey;autoIncrement;column:item_set_id" json:"itemset_id"`
	SalesProjectItemSetContent
}

func (SalesProjectItemSet) TableName() string {
	return "tbl_trans_sales_project_item_set"
}

type SalesProjectItemSetAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefID uint `json:"ref_id"`
	SalesProjectItemSetContent
	At
}

func (SalesProjectItemSetAt) TableName() string {
	return "z_tbl_trans_sales_project_item_set_at"
}
