package models

import "time"

type DeliveryReceiptContent struct {
	ID             uint                       `gorm:"primaryKey"`
	DeliveryNumber string                     `gorm:"size:50;unique;not null"`
	SalesOrderID   uint                       `json:"sales_order_id"`
	ReleasedByID   uint                       `json:"released_by_id"`
	DeliveredByID  uint                       `json:"deliverd_by_id"`
	ReceiptNumber  string                     `json:"receipt_number"`
	RecipientName  string                     `gorm:"size:150" json:"recipient_name"`
	DeliveryDate   time.Time                  `json:"delivery_date"`
	DeliveryCost   float64                    `json:"delivery_cost"`
	Remarks        string                     `gorm:"size:255" json:"remarks"`
	Items          []DeliveryReceiptItemModel `gorm:"foreignKey:DeliveryReceiptID; constraint:OnDelete:CASCADE" json:"items"`
	TripCost       TripCostModel              `gorm:"foreignKey:DeliveryReceiptID; references:ID; constraint:OnDelete:CASCADE" json:"trip_cost"`
	//	Attachments    []Attachment
}

type DeliveryReceiptModel struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeliveryReceiptContent
}

func (DeliveryReceiptModel) TableName() string {
	return "tbl_delivery_receipt"
}

type DeliveryReceiptAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DeliveryReceiptContent
	At
}

func (DeliveryReceiptAt) TableName() string {
	return "z_tbl_delivery_receipt"
}

type DeliveryReceiptItemContent struct {
	ID                uint `gorm:"primaryKey"`
	DeliveryReceiptID uint `json:"delivery_receipt_id"`
	SalesOrderItemID  uint `json:"delivery_item_id"`
	QuantityDelivered int  `json:"quantity_delivered"`
}

type DeliveryReceiptItemModel struct {
	ID uint `gorm:"primaryKey"`
	DeliveryReceiptItemContent
}

func (DeliveryReceiptItemModel) TableName() string {
	return "tbl_delivery_receipt_item"
}

type DeliveryReceiptItemAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DeliveryReceiptItemContent
	At
}

func (DeliveryReceiptItemAt) TableName() string {
	return "z_tbl_delivery_receipt_item"
}

type TripCostContent struct {
	ID                uint    `gorm:"primaryKey"`
	DeliveryReceiptID uint    `json:"delivery_receipt_id"`
	Vehicle           string  `gorm:"size:100" json:"vehicle"`
	DriverName        string  `gorm:"size:150" json:"driver_name"`
	FuelCost          float64 `json:"fuel_cost"`
	TollFee           float64 `json:"total_fee"`
	OtherExpenses     float64 `json:"other_expenses"`
	TotalCost         float64 `json:"total_cost"`
}

type TripCostModel struct {
	ID uint `gorm:"primaryKey"`
	TripCostContent
}

func (TripCostModel) TableName() string {
	return "tbl_trip_cost"
}

type TripCostContentAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	TripCostContent
	At
}

func (TripCostContentAt) TableName() string {
	return "z_tbl_trip_cost"
}
