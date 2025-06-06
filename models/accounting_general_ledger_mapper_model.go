package models

type GeneralLedgerMapperPayload struct {
	Payload []GeneralLedgerMapper `json:"Payload"`
}

type GeneralLedgerMapper struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	PseudoAccount string `gorm:"unique" json:"pseudo_account"`
	AccountId     uint   `json:"account_id"`
}

func (GeneralLedgerMapper) TableName() string {
	return "tbl_setup_chart_of_account_mapping"
}

type GeneralLedgerMapperAt struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	RefId         uint   `json:"ref_id"`
	PseudoAccount string `gorm:"unique" json:"pseudo_account"`
	AccountId     uint   `json:"account_id"`
	At
}

func (GeneralLedgerMapperAt) TableName() string {
	return "z_tbl_setup_chart_of_account_mapping_at"
}
