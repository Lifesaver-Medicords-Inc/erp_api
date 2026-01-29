package accounting_models

type InvoicePOView struct {
	ID           uint   `json:"id"`
	PONumber     string `json:"po_number"`
	DocDate      string `json:"doc_date"`
	SupplierName string `json:"supplier_name"`
}

type InvoicePODetailView struct {
	PurchaseOrderDetailsId uint    `json:"purchase_order_details_id"`
	BasedId                uint    `json:"based_id"`
	ItemCode               string  `json:"item_code"`
	ItemDescription        string  `json:"item_description"`
	OrderUom               string  `json:"order_uom"`
	OrderQty               uint    `json:"order_qty"`
	ReqUom                 string  `json:"req_uom"`
	ReqQty                 uint    `json:"req_qty"`
	UnitPrice              float64 `json:"unit_price"`
	TotalCost              float64 `json:"total_cost"`
}
