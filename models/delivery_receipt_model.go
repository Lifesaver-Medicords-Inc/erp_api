package models

import (
	"time"

	"gorm.io/gorm"
)

type DeliveryReceiptContent struct {
	DeliveryReceiptNumber string             `gorm:"uniqueIndex;size:50;not null" json:"delivery_receipt_number"`
	DeliveryReference     string             `gorm:"size:50" json:"delivery_reference,omitempty"`
	OrderID               uint               `gorm:"not null" json:"order_id"`
	DeliveryDate          time.Time          `gorm:"not null" json:"delivery_date"`
	DriverName            string             `gorm:"size:100" json:"driver_name,omitempty"`
	PlateNumber           string             `gorm:"size:50" json:"plate_number,omitempty"`
	Status                string             `gorm:"size:50;default:'In Transit'" json:"status"`
	Remarks               string             `gorm:"size:255" json:"remarks,omitempty"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	DeletedAt             gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	ReleaseID             uint               ` json:"release_id"`
	DeliveryType          string             `json:"delivery_type,omitempty"`
	Order                 *Order             `gorm:"foreignKey:Order_ID;references:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
	ItemReleases          []ItemReleaseModel `gorm:"foreignKey:ID;references:ReleaseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item_releases,omitempty"`
	TripCost              *TripCostModel     `gorm:"foreignKey:DeliveryReceiptID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"trip_cost,omitempty"`
	AttachedFile          *ReceiptFileModel  `gorm:"foreignKey:DeliveryReceiptID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"attached_file,omitempty"`
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

type TripCostContent struct {
	DeliveryReceiptID uint                  `gorm:"not null;uniqueIndex" json:"delivery_receipt_id"`
	CostType          string                `gorm:"default:0" json:"cost_type"`
	TotalCost         float64               `gorm:"default:0" json:"total_cost"`
	Remarks           string                `gorm:"size:255" json:"remarks,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	DeletedAt         gorm.DeletedAt        `gorm:"index" json:"deleted_at,omitempty"`
	DeliveryReceipt   *DeliveryReceiptModel `gorm:"foreignKey:DeliveryReceiptID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"delivery_receipt,omitempty"`
	VehiclePlateNo    string                `gorm:"size:100" json:"vehicle_plate_no"`
	DriverName        string                `gorm:"size:150" json:"driver_name"`
	OtherExpenses     float64               `json:"other_expenses"`
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

type ReceiptFileContent struct {
	FileName          string                `json:"file_name"`
	OriginalName      string                `json:"original_name"`
	FilePath          string                `json:"file_path"`
	Type              string                `json:"type"`
	Size              int                   `json:"size"`
	DeliveryReceiptID uint                  `json:"delivery_receipt_id"`
	Receipt           *DeliveryReceiptModel `gorm:"foreignKey:DeliveryReceiptID;references:ID;" json:"receipt,omitempty"`
}

type ReceiptFileModel struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ReceiptFileContent
}

func (ReceiptFileModel) TableName() string {
	return "tbl_receipt_file"
}

type ReceiptFileAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	At
}

func (ReceiptFileAt) TableName() string {
	return "z_tbl_receipt_file_at"
}
