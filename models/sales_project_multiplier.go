package models

type SalesProjectMultiplierContent struct {
	// Parent ID references to Sales Quotation Model
	BasedId     uint   `json:"based_id"`
	Brand       string `json:"brand"`
	Component   string `json:"component"`
	Description string `json:"description"`
	Multiplier  string `json:"multiplier"`
}

type SalesProjectMultiplier struct {
	MultiplierID uint `gorm:"primarykey" json:"multiplier_id"`
	SalesProjectMultiplierContent
}

func (SalesProjectMultiplier) TableName() string {
	return "tbl_trans_sales_project_multiplier"
}

type SalesProjectMultiplierAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesProjectMultiplierContent
	At
}

func (SalesProjectMultiplierAt) TableName() string {
	return "z_tbl_trans_sales_project_multiplier_at"
}
