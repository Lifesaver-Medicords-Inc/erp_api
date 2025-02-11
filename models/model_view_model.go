package models

type ModelView struct {
	Model
	RelatedName  string `json:"related_name"`
	RelatedBrand string `json:"related_brand"`
}

func (ModelView) TableName() string {
	return "vw_item_model_list"
}
