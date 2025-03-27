package models

type BpiSuppliersViewContent struct {
	BasedId      uint    `json:"based_id"`
	SupplierCode string  `json:"supplier_code"`
	ItemId       uint    `json:"item_id"`
	ItemName     string  `json:"item_name"`
	Price        float64 `json:"price"`
}

type BpiSuppliersView struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiSuppliersViewContent
}

func (BpiSuppliersView) TableName() string {
	return "GetBpiSuppliers"
}
