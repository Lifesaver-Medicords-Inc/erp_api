package models

type ItemView struct {
	Item
	ItemName       string `json:"item_name"`
	ItemModel      string `json:"item_model"`
	CatalogueYear  string `json:"catalogue_year"`
	ItemClass      string `json:"item_class"`
	ItemBrand      string `json:"item_brand"`
	UnitOfMeasure  string `json:"unit_of_measure"`
	TradeTypeId    string `json:"trade_type_id"`
	TradeTypeNames string `json:"trade_type_names"`

	// TradeStatus   string `json:"trade_status"`

}

func (ItemView) TableName() string {
	return "vw_items"
}
