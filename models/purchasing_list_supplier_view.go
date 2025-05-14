package models

type PurchasingListSupplierViewContent struct {
	SupplierId int    `json:"supplier_id"`
	Supplier   string `json:"supplier"`
	ContactNo  string `json:"contact_no"`
}

type PurchasingListSupplierView struct {
	PurchasingListSupplierViewContent
}

func (PurchasingListSupplierView) TableName() string {
	return "vw_get_purchasing_list_supplier"
}
