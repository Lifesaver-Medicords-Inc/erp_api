package models

type PurchaseOrderContent struct {
	SupplierId     int     `json:"supplier_id"`
	SupplierName   string  `json:"supplier_name"`
	SupplierCode   string  `json:"supplier_code"`
	Address        string  `json:"address"`
	TinNo          string  `json:"tin_no"`
	FaxNo          string  `json:"fax_no"`
	MainTelNo      string  `json:"main_tel_no"`
	ShipTypeId     int     `json:"ship_type_id"`
	DeliverTo      string  `json:"deliver_to"`
	DeliverVia     string  `json:"deliver_via"`
	DocNo          string  `json:"doc_no"`
	Date           string  `json:"date"`
	OrderType      string  `json:"order_type"`
	TaxCode        string  `json:"tax_code"`
	TaxCodePercent string  `json:"tax_code_percent"`
	PaymentTermsId int     `json:"payment_terms_id"`
	BillTo         string  `json:"bill_to"`
	RefDocNo       string  `json:"ref_doc_no"`
	Status         string  `json:"status"`
	Remarks        string  `json:"remarks"`
	NetOfVat       float64 `json:"net_of_vat"`
	Vat            float64 `json:"vat"`
	TotalAmountDue float64 `json:"total_amount_due"`
}

type PurchaseOrder struct {
	ID uint `gorm:"primarykey" json:"id"`
	PurchaseOrderContent
}

func (PurchaseOrder) TableName() string {
	return "tbl_purchasing_purchase_order"
}

type PurchaseOrderAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchaseOrderContent
	At
}

func (PurchaseOrderAt) TableName() string {
	return "z_tbl_purchasing_purchase_order_at"
}
