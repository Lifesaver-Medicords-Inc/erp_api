package dispatching_models

import "github.com/pierceperado/smpc/models"

type ReceiptFileContent struct {
	DeliveryReceiptID     uint   `json:"delivery_receipt_id"`
	DeliveryReceiptCostID uint   `json:"delivery_receipt_cost_id"`
	FileName              string `json:"file_name"`
	OriginalName          string `json:"original_name"`
	FilePath              string `json:"file_path"`
	Type                  string `json:"type"`
	Size                  int    `json:"size"`
}

type ReceiptFile struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ReceiptFileContent
}

func (ReceiptFile) TableName() string {
	return "tbl_receipt_file"
}

type ReceiptFileAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	models.At
}

func (ReceiptFileAt) TableName() string {
	return "z_tbl_receipt_file_at"
}
