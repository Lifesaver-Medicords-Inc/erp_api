package models

type TradeTypeContent struct {
	BasedId uint   `json:"based_id"`
	Value   string `json:"value"`
}

type TradeType struct {
	ID uint `gorm:"primarykey" json:"id"`
	TradeTypeContent
}

func (TradeType) TableName() string {
	return "tbl_item_trade_type"
}

type TradeTypeAt struct {
	ID uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	TradeTypeContent
	At
}

func (TradeTypeContent) TableName() string {
	return "z_tbl_item_trade_type_at"
}