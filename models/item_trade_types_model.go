package models

type ItemTradeTypeContent struct {
	ItemId      uint `json:"item_id"`
	TradeTypeId uint `json:"trade_type_id"`
}

type ItemTradeType struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemTradeTypeContent
}

func (ItemTradeType) TableName() string {
	return "tbl_setup_item_trade_type"
}

type ItemTradeTypeAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemTradeTypeContent
	At
}

func (ItemTradeTypeAt) TableName() string {
	return "z_tbl_setup_item_trade_type_at"
}
