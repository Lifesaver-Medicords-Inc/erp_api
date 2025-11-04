package models

import (
	"time"

	"gorm.io/gorm"
)

type SalesOrderContent struct {
	CustomerName     string                 `gorm:"size:150" json:"customer_name"`
	TotalCost        float64                `gorm:"default:0" json:"total_cost"`
	Code             string                 `json:"code"`
	DeliverTo        string                 `json:"deliver_to"`
	BillTo           string                 `json:"bill_to"`
	Tin              string                 `json:"tin"`
	Receiver         string                 `json:"receiver"`
	ContactNo        string                 `json:"contact_no"`
	DocNo            string                 `json:"doc_no"`
	ReferenceDocCode string                 `json:"reference_doc_code"`
	SalesExecutive   string                 `json:"sales_executive"`
	DeliveryStatus   string                 `json:"delivery_status"`
	DeliveryDate     time.Time              `json:"delivery_date"`
	SalesOrderNumber string                 `gorm:"uniqueIndex;size:50;not null" json:"sales_order_number"`
	CustomerID       uint                   `gorm:"not null" json:"customer_id"`
	OrderDate        time.Time              `gorm:"not null" json:"order_date"`
	Status           string                 `gorm:"size:50;default:'Pending'" json:"status"`
	Remarks          string                 `gorm:"size:255" json:"remarks,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`
	Items            []SalesOrderItemModel  `gorm:"foreignKey:SalesOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items,omitempty"`
	DeliveryReceipts []DeliveryReceiptModel `gorm:"foreignKey:SalesOrderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"delivery_receipts,omitempty"`
	//Attachments     []Attachment
}

type SalesOrderModel struct {
	ID uint `gorm:"primaryKey"  json:"id"`
	SalesOrderContent
}

func (SalesOrderModel) TableName() string {
	return "tbl_sales_order"
}

type SalesOrderAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesOrderContent
	At
}

func (SalesOrderAt) TableName() string {
	return "z_tbl_sales_order"
}

type SalesOrderItemContent struct {
	Status          string             `json:"status"`
	SalesOrderID    uint               `gorm:"not null;index" json:"sales_order_id"`
	ItemID          uint               `gorm:"not null;index" json:"item_id"`
	QuantityOrdered float64            `gorm:"not null" json:"quantity_ordered"`
	UnitPrice       float64            `gorm:"not null" json:"unit_price"`
	Remarks         string             `gorm:"size:255" json:"remarks,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	SalesOrder      *SalesOrderModel   `gorm:"foreignKey:SalesOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sales_order,omitempty"`
	Item            *Item              `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item,omitempty"`
	Releases        []ItemReleaseModel `gorm:"foreignKey:SalesOrderItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"releases,omitempty"`
}

type SalesOrderItemModel struct {
	ID uint `gorm:"primaryKey"  json:"id"`
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
