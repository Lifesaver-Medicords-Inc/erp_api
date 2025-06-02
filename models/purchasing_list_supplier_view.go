package models

type PurchasingListSupplierViewContent struct {
	SupplierId     int    `json:"supplier_id"`
	SupplierCode   string `json:"supplier_code"`
	SupplierName   string `json:"supplier_name"`
	ContactNo      string `json:"contact_no"`
	Address        string `json:"address"`
	TinNo          string `json:"tin_no"`
	FaxNo          string `json:"fax_no"`
	MainTelNo      string `json:"main_tel_no"`
	TaxCode        string `json:"tax_code"`
	PaymentTermsId int    `json:"payment_terms_id"`
	ItemIds        string `json:"item_ids"`
}

type PurchasingListSupplierView struct {
	PurchasingListSupplierViewContent
}

func (PurchasingListSupplierView) TableName() string {
	return "vw_get_purchasing_list_supplier"
}
