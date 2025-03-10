package models

type PurchasingListViewContent struct {
	OrderDetailIds  string `json:"order_detail_ids"`
	BasedIds        string `json:"based_ids"`
	SalesOrderNos   string `json:"sales_order_nos"`
	ItemId          uint   `json:"item_id"`
	ItemCode        string `json:"item_code"`
	ItemName        string `json:"item_name"`
	ItemDescription string `json:"item_description"`
	UnitPrices      string `json:"unit_prices"`
	Qtys            string `json:"qtys"`
	UnitOfMeasures  string `json:"unit_of_measures"`
	TotalQty        uint   `json:"total_qty"`
}

type PurchasingListView struct {
	PurchasingListViewContent
}

func (PurchasingListView) TableName() string {
	return "vw_get_purchasing_list"
}
