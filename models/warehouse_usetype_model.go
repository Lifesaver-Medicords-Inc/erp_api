package models

type UseTypeContent struct {
	Name string `json:"name"`
}

type UseType struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;not null" json:"code"`
	UseTypeContent
}

func (UseType) TableName() string {
	return "tbl_setup_usetype"
}

type UseTypeAt struct {
	ID    uint   `gorm:"primary" json:"id"`
	RefID uint   `json:"ref_id"`
	Code  string `json:"code"`
	UseTypeContent
	At
}

func (UseTypeAt) TableName() string {
	return "z_tbl_setup_usetype_at"
}
