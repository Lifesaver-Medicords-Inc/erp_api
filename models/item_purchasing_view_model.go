package models

type ItemPurchasingView struct {
	ItemPurchasing
	SupplierName     string `json:"supplier_name"`
	PaymentTermsName string `json:"payment_terms_name"`
	SupplierTypeName string `json:"supplier_type_name"`
}

func (ItemPurchasingView) TableName() string {
	return "vw_item_purchasing_list"
}
