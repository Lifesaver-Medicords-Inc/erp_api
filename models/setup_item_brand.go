package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Brand struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	At   `gorm:"-"`
}

func (Brand) TableName() string {
	return "tbl_setup_item_brand"
}

func (b *Brand) AfterSave(tx *gorm.DB) (err error) {
	fmt.Println("Brand Result:", b)
	brandAt := BrandAt{
		Brand: *b,
		At:    b.At,
	}

	if err := tx.Create(&brandAt).Error; err != nil {
		return err
	}

	return nil
}

type BrandAt struct {
	Brand
	At
}

func (BrandAt) TableName() string {
	return "z_tbl_setup_item_brand_at"
}
