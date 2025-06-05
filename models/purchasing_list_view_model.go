package models

type SOPurchasingListView struct {
	ItemId          uint   `json:"item_id"`
	Purchaser       string `json:"purchaser"`
	OrderIds        string `json:"order_ids"`
	OrderDetailIds  string `json:"order_detail_ids"`
	SalesOrderNos   string `json:"sales_order_nos"`
	ProjectNames    string `json:"project_names"`
	SalesExecutives string `json:"sales_executives"`
	Discounts       string `json:"discounts"`
	UnitPrices      string `json:"unit_prices"`
	QuoteSupplier   string `json:"quote_supplier"`
	ItemCode        string `json:"item_code"`
	ItemDescription string `json:"item_description"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	ItemName        string `json:"item_name"`
	ItemBrand       string `json:"item_brand"`
	CommitmentDates string `json:"commitment_dates"`
	Qtys            string `json:"qtys"`
	TotalQty        uint   `json:"total_qty"`
}

func (SOPurchasingListView) TableName() string {
	return "vw_get_purchasing_so_purchase_list"
}
