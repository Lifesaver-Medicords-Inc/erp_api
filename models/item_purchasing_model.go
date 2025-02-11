package models

type ItemPurchasingContent struct {
	BasedId          uint    `json:"based_id"`
	SupplierNameId   uint    `json:"supplier_name_id"`
	PaymentTermsId   uint    `json:"payment_terms_id"`
	Price            float64 `json:"price"`
	SupplierTypeName string  `json:"supplier_type_name"`
	ValidityPeriod   string  `json:"validity_period"`
}

type ItemPurchasing struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemPurchasingContent
}

func (ItemPurchasing) TableName() string {
	return "tbl_setup_item_purchasing"
}

type ItemPurchasingAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemPurchasingContent
	At
}

func (ItemPurchasingAt) TableName() string {
	return "z_tbl_setup_item_purchasing_at"
}
