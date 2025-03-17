package models

type BomViewItemList struct {
	Id        uint   `gorm:"primarykey" json:"id"`
	ItemBomID uint   `json:"item_bom_id"`
	ItemID    uint   `json:"item_id"`
	Size      string `json:"size"`
	BomQty    uint   `json:"bom_qty"`
	ItemCode  string `json:"item_code"`
	ShortDesc string `json:"short_desc"`
	UomName   string `json:"uom_name"`
	UnitPrice uint   `json:"unit_price"`
}

func (BomViewItemList) TableName() string {
	return "vw_item_bom_list"
}
