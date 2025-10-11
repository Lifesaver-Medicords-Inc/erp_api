package models

import "time"

type ItemReleaseDetailContent struct {
	ItemReleaseID    uint `json:"item_release_id"`
	SalesOrderItemID uint `json:"sales_order_item_id"`
	Quantity         int  `json:"quantity"`
}

type ItemReleaseDetailModel struct {
	ID uint `gorm:"primaryKey"`
	ItemReleaseDetailContent
}

func (ItemReleaseDetailModel) TableName() string {
	return "tbl_item_release_detail"
}

type ItemReleaseDetailAt struct {
	ID uint `gorm:"primaryKey"`
	ItemReleaseDetailContent
}

func (ItemReleaseDetailAt) TableName() string {
	return "z_item_release_detail"
}

type ItemReleaseContent struct {
	SalesOrderID  uint                     `json:"sales_order_id"`
	RequestedByID uint                     `json:"requested_by_id"`
	ApprovedByID  *uint                    `json:"approved_by_id"`
	Status        string                   `gorm:"size:50" json:"status"`
	CreatedAt     time.Time                `json:"created_by_id"`
	ReleasedAt    *time.Time               `json:"release_at"`
	Items         []ItemReleaseDetailModel `gorm:"foreignKey:ItemReleaseID; constraint:OnDelete:CASCADE" json:"items"`
}

type ItemReleaseModel struct {
	ID uint `gorm:"primaryKey"`
	ItemReleaseContent
}

func (ItemReleaseModel) TableName() string {
	return "tbl_item_release"
}

type ItemReleaseAt struct {
	ID uint `gorm:"primaryKey"`
	ItemReleaseContent
	At
}

func (ItemReleaseAt) TableName() string {
	return "z_tbl_item_release"
}
