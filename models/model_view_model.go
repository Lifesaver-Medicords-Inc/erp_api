package models

type ModelView struct {
	Model
	RelatedName  string `json:"related_name" gorm:"->"`
	RelatedBrand string `json:"related_brand" gorm:"->"`
}

func (ModelView) TableName() string {
	return "ItemModelList"
}
