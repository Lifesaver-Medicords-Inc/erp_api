package models

import "time"

type SalesOrderContent struct {
	OrderNumber     string  `gorm:"size:50;unique;not null"`
	CustomerName    string  `gorm:"size:150"`
	Status          string  `gorm:"size:50;default:'Pending'"`
	TotalCost       float64 `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SalesOrderItems []SalesOrderItemModel `gorm:"foreignKey:SalesOrderID; constraint:OnDelete:CASCADE"`
	//Attachments     []Attachment
}

type SalesOrderModel struct {
	ID uint `gorm:"primaryKey"`
	SalesOrderContent
}

func (SalesOrderModel) TableName() string {
	return "tbl_sales_order"
}

type SalesOrderAt struct {
	ID    uint `gorm:"primaryKey"`
	RefId uint `json:"ref_id"`
	SalesOrderContent
	At
}

func (SalesOrderAt) TableName() string {
	return "z_tbl_sales_order"
}

type SalesOrderItemContent struct {
	SalesOrderID uint
	ItemName     string `gorm:"size:150"`
	Quantity     int
	DeliveredQty int
	UnitPrice    float64
	TotalAmount  float64
}

type SalesOrderItemModel struct {
	ID uint `gorm:"primaryKey"`
	SalesOrderItemContent
}

func (SalesOrderItemModel) TableName() string {
	return "tbl_sales_order_item"
}

type SalesOrderItemAt struct {
	ID    uint `gorm:"primaryKey"`
	RefId uint `json:"ref_id"`
	SalesOrderItemContent
	At
}

func (SalesOrderItemAt) TableName() string {
	return "z_tbl_sales_order_item"
}
