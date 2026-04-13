package inventory_models

type PurchaseOrderReceivingView struct {
	PurchaseOrderId int    `json:"purchase_order_id"`
	SupplierId      int    `json:"supplier_id"`
	Supplier        string `json:"supplier"`
	SupplierCode    string `json:"supplier_code"`
}

type PurchaseOrderReceivingDetailsView struct {
	PurchaseOrderDetailsId int    `json:"purchase_order_details_id"`
	ItemId                 int    `json:"item_id"`
	ItemCode               string `json:"item_code"`
	ItemDesc               string `json:"item_desc"`
	OrderedQty             int    `json:"ordered_qty"`
	OrderedUom             string `json:"ordered_uom"`
	RemainingQty           int    `json:"remaining_qty"`
	RemainingUom           string `json:"remaining_uom"`
}

type PurchaseOrderDocView struct {
	PurchaseOrderId uint   `json:"purchase_order_id"`
	PoDocNo         string `json:"po_doc_no"`
}

func (PurchaseOrderDocView) TableName() string {
	return "vw_get_purchase_order_doc"
}
