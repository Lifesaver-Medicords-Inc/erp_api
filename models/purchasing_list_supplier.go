package models

type PurchasingListSupplierViewContent struct {
	BpiId            int     `json:"bpi_id"`
	BpiItemId           int     `json:"bpi_item_id"`
	ItemId           int     `json:"item_id"`
	CurrentListPrice float64 `json:"current_list_price"`
	PaymentTermsId   int     `json:"payment_terms_id"`
	SupplierName     string  `json:"supplier_name"`
	ContactNo        string  `json:"contact_no"`
}

type PurchasingListSupplierView struct {
	PurchasingListSupplierViewContent
}

func (PurchasingListSupplierView) TableName() string {
	return "vw_get_purchasing_list_supplier"
}
