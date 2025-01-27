package models

type EntityContent struct {
	Name string `json:"name"`
}

type Entity struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	EntityContent
}

func (Entity) TableName() string {
	return "tbl_setup_bpi_entity"
}

type EntityAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	EntityContent
	At
}

func (EntityAt) TableName() string {
	return "z_tbl_setup_bpi_entity_at"
}
