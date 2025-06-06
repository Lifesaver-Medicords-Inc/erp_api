package models

type PRPurchasingListView struct {
	ItemId                       uint   `json:"item_id"`
	Purchaser                    string `json:"purchaser"`
	PurchaseRequisitionDetailIds string `json:"purchase_requisition_detail_ids,"`
	PurchaseRequisitionIds       string `json:"purchase_requisition_ids"`
	PurchaseRequisitionNos       string `json:"purchase_requisition_nos"`
	Requestors                   string `json:"requestors"`
	Departments                  string `json:"departments"`
	ItemCode                     string `json:"item_code"`
	ItemDescription              string `json:"item_description"`
	UnitOfMeasure                string `json:"unit_of_measure"`
	ItemName                     string `json:"item_name"`
	ItemBrand                    string `json:"item_brand"`
	CommitmentDates              string `json:"commitment_dates"`
	Qtys                         string `json:"qtys"`
	TotalQty                     uint   `json:"total_qty"`
	// Discounts              string `json:"discounts"`
	// UnitPrices             string `json:"unit_prices"`
	// QuoteSupplier          string `json:"quote_supplier"`
}

func (PRPurchasingListView) TableName() string {
	return "vw_get_purchasing_pr_purchase_list"
}
