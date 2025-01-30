package models

type BpiEntityContent struct {
	BpiGeneralId uint `json:"bpi_general_id"`
	EntityId     uint `json:"entity_id"`
}

type BpiEntity struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiEntityContent
}

func (BpiEntity) TableName() string {
	return "tbl_bpi_entity"
}

type BpiEntityAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiEntityContent
	At
}

func (BpiEntityAt) TableName() string {
	return "z_tbl_bpi_entity_at"
}
