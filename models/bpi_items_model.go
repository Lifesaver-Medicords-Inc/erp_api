package models

type BpiItemContent struct {
	BasedId        uint    `json:"based_id"`
	BranchId       uint    `json:"branch_id"`
	PaymentTermsId uint    `json:"payment_terms_id"`
	ItemAccountId  uint    `json:"item_account_id"`
	ItemId         uint    `json:"item_id"`
	TaxCode        string  `json:"tax_code"`
	ItemTaxCode    string  `json:"item_tax_code"`
	Price          float64 `json:"price"`
	Notes          string  `json:"notes"`
}

type BpiItems struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiItemContent
}

func (BpiItems) TableName() string {
	return "tbl_bpi_items"
}

type BpiItemsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiItemContent
	At
}

func (BpiItemsAt) TableName() string {
	return "z_tbl_bpi_items_at"
}
