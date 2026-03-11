package dispatching_models

import "github.com/pierceperado/smpc/models"

type DeliveryReceiptItemsContent struct {
	DeliveryReceiptID uint   `gorm:"not null;index" json:"delivery_receipt_id"`
	ItemID            uint   `gorm:"not null" json:"item_id"`
	Qty               int    `gorm:"not null" json:"qty"`
	UnitOfMeasure     string `gorm:"not null" json:"unit_of_measure"`
	ItemCode          string `gorm:"size:50" json:"item_code,omitempty"`
	ItemDescription   string `gorm:"size:255" json:"item_description,omitempty"`
	SerialNo          string `gorm:"size:100" json:"serial_no,omitempty"`
}

type DeliveryReceiptItems struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeliveryReceiptItemsContent
}

func (DeliveryReceiptItems) TableName() string {
	return "tbl_dispatching_delivery_receipt_items"
}

type DeliveryReceiptItemsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DeliveryReceiptItemsContent
	models.At
}

func (DeliveryReceiptItemsAt) TableName() string {
	return "z_tbl_dispatching_delivery_receipt_items_at"
}
