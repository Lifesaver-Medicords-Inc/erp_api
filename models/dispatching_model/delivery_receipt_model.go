package dispatching_models

import "github.com/pierceperado/smpc/models"

type DeliveryReceiptContent struct {
	CustomerID           uint                   `gorm:"not null" json:"customer_id"`
	CustomerName         string                 `gorm:"size:255" json:"customer_name,omitempty"`
	CustomerCode         string                 `gorm:"size:50" json:"customer_code,omitempty"`
	Address              string                 `gorm:"size:255" json:"address,omitempty"`
	TinNo                string                 `gorm:"size:50" json:"tin_no,omitempty"`
	ShipTypeID           uint                   `json:"ship_type_id,omitempty"`
	DeliverTo            string                 `gorm:"size:255" json:"deliver_to,omitempty"`
	ShipVia              string                 `json:"ship_via,omitempty"`
	Att                  string                 `gorm:"type:text" json:"att,omitempty"`
	Date                 string                 `json:"date,omitempty"`
	DeliveryDate         string                 `json:"delivery_date,omitempty"`
	SalesOrderID         uint                   `gorm:"not null" json:"sales_order_id"`
	ItemReleaseID        uint                   `gorm:"not null" json:"item_release_id"`
	SalesExecutive       string                 `gorm:"not null" json:"sales_executive"`
	DeliveryReceiptItems []DeliveryReceiptItems `gorm:"foreignKey:DeliveryReceiptID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"delivery_receipt_items,omitempty"`
	DeliveryReceiptCosts []DeliveryReceiptCosts `gorm:"foreignKey:DeliveryReceiptID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"delivery_receipt_costs,omitempty"`
}

type DeliveryReceipt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"not null;uniqueIndex:idx_tbl_dispatching_delivery_receipt_doc_no" json:"doc_no"`
	DeliveryReceiptContent
}

func (DeliveryReceipt) TableName() string {
	return "tbl_dispatching_delivery_receipt"
}

type DeliveryReceiptAt struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	RefId uint   `json:"ref_id"`
	DocNo string `gorm:"size:50;index"`
	DeliveryReceiptContent
	models.At
}

func (DeliveryReceiptAt) TableName() string {
	return "z_dispatching_tbl_delivery_receipt"
}
