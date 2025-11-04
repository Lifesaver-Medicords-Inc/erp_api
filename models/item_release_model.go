package models

import (
	"time"

	"gorm.io/gorm"
)

type ItemReleaseContent struct {
	RequestedByID     uint                  `json:"requested_by_id"`
	RequestedByName   string                `json:"requested_by_name"`
	ApprovedByID      *uint                 `json:"approved_by_id"`
	ApprovedByName    string                `json:"approved_by_name"`
	ReleaseByID       string                `json:"release_by_id"`
	ReleaseByName     string                `json:"release_by_name"`
	SalesOrderItemID  uint                  `gorm:"not null;index" json:"sales_order_item_id"`
	DeliveryReceiptID uint                  `gorm:"not null;index" json:"delivery_receipt_id"`
	QuantityReleased  float64               `gorm:"not null" json:"quantity_released"`
	SerialNumber      string                `gorm:"size:100" json:"serial_number,omitempty"`
	DepartedAt        *time.Time            `json:"departed_at,omitempty"`
	ArrivedAt         *time.Time            `json:"arrived_at,omitempty"`
	ReturnedAt        *time.Time            `json:"returned_at,omitempty"`
	Status            string                `gorm:"size:50;default:'Pending'" json:"status"`
	Remarks           string                `gorm:"size:255" json:"remarks,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	DeletedAt         gorm.DeletedAt        `gorm:"index" json:"deleted_at,omitempty"`
	SalesOrderItem    *SalesOrderItemModel  `gorm:"foreignKey:SalesOrderItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sales_order_item,omitempty"`
	DeliveryReceipt   *DeliveryReceiptModel `gorm:"foreignKey:DeliveryReceiptID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"delivery_receipt,omitempty"`
	VehicleID         uint                  `json:"vehicle_id"`
	Vehicle           *VehicleModel         `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
}

type ItemReleaseModel struct {
	ID uint `gorm:"primaryKey"  json:"id"`
	ItemReleaseContent
}

func (ItemReleaseModel) TableName() string {
	return "tbl_item_release"
}

type ItemReleaseAt struct {
	ID    uint `gorm:"primaryKey"  json:"id"`
	RefId uint `json:"ref_id"`
	ItemReleaseContent
	At
}

func (ItemReleaseAt) TableName() string {
	return "z_tbl_item_release"
}
