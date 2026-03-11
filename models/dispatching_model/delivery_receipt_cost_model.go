package dispatching_models

import "github.com/pierceperado/smpc/models"

type DeliveryReceiptCostContent struct {
	DeliveryReceiptID uint          `gorm:"not null;index" json:"delivery_receipt_id"`
	CostTypeID        uint          `gorm:"default:0" json:"cost_type_id"`
	Description       string        `gorm:"size:255" json:"description,omitempty"`
	Amount            float64       `gorm:"default:0" json:"amount"`
	Multiplier        float64       `gorm:"default:0" json:"multiplier"`
	TotalCost         float64       `gorm:"default:0" json:"total_cost"`
	ReceiptFiles      []ReceiptFile `gorm:"foreignKey:DeliveryReceiptCostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"receipt_files,omitempty"`
}
type DeliveryReceiptCosts struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeliveryReceiptCostContent
}

func (DeliveryReceiptCosts) TableName() string {
	return "dispatching_tbl_delivery_receipt_costs"
}

type DeliveryReceiptCostsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DeliveryReceiptCostContent
	models.At
}

func (DeliveryReceiptCostsAt) TableName() string {
	return "z_dispatching_tbl_delivery_receipt_costs"
}
