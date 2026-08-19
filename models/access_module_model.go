package models

// AccessModuleModel is the master catalog of every grantable screen/action across all
// SMPC apps (Admin, Sales, Inventory, Engineering, Dispatching, Accounting) - sourced
// from SMPC_User_Access_Level_List.xlsx and seeded once via SeedAccessModules (see
// initializers/seed_access_modules.go). tbl_position_access (unchanged) grants a
// Position one of these Codes; this table is just the read-only list of what Codes
// exist and how they're organized, so the Admin Access Control screen has something
// real to render instead of a hardcoded 5-item list.
//
// Code is deliberately NOT globally unique - two different apps can legitimately share
// the same Module/Submodule text (e.g. every app has its own "Application Shell.Login"),
// and each app's copy needs to be granted independently. Uniqueness is enforced in
// application code (SeedAccessModules checks AppName+Code+Button before inserting), not
// via a DB constraint - a composite index/unique constraint across two long nvarchar
// columns risks exceeding SQL Server's 900-byte max index key length once Code gets
// long (deeply nested Module.Submodule.Button text), so this deliberately has no index.
// Explicit sizes below keep every column safely under that ceiling regardless.
type AccessModuleContent struct {
	ID          uint   `gorm:"primarykey; autoIncrement" json:"id"`
	AppName     string `gorm:"not null; size:100" json:"app_name"`
	Module      string `gorm:"not null; size:150" json:"module"`
	Submodule   string `gorm:"not null; size:150" json:"submodule"`
	Button      string `gorm:"size:150" json:"button"` // empty for a page-level (Kind="page") entry
	Code        string `gorm:"not null; size:400" json:"code"`
	Kind        string `gorm:"not null; size:20" json:"kind"` // "page" | "button"
	IsSensitive bool   `json:"is_sensitive"`                  // Delete/Approve/Reject/Void/Finalize/etc - flagged for later enforcement, not enforced yet
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

type AccessModuleModel struct {
	AccessModuleContent
}

func (AccessModuleModel) TableName() string {
	return "tbl_access_modules"
}
