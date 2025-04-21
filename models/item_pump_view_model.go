package models

type ItemPumpSpecsViewContent struct {
	ItemId       uint   `json:"item_id"`
	ItemNameId   uint   `json:"item_name_id"`
	ItemName     string `json:"item_name"`
	TemplateName string `json:"template_name"`
	ItemTitle    string `json:"item_title"`
	ItemValue    string `json:"item_value"`
}

type ItemPumpSpecsView struct {
	ItemPumpSpecsViewContent
}

func (ItemPumpSpecsView) TableName() string {
	return "vw_PumpSpecifications"
}
