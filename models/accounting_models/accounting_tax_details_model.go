package accounting_models

import "github.com/pierceperado/smpc/models"

type TaxDetailsContent struct {
	TaxCodeId uint    `json:"tax_code_id"`
	ValidFrom string  `json:"valid_from"`
	ValidTo   string  `json:"valid_to"`
	TaxRate   float64 `json:"tax_rate"`
}
type TaxDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	TaxDetailsContent
}

func (TaxDetails) TableName() string {
	return "tbl_setup_tax_details"
}

type TaxDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	TaxDetailsContent
	models.At
}

func (TaxDetailsAt) TableName() string {
	return "z_tbl_setup_tax_details_at"
}

type TaxDetailsView struct {
	Id          int     `json:"id"`
	TaxCodeId   uint    `json:"tax_code_id"`
	ValidFrom   string  `json:"valid_from"`
	ValidTo     string  `json:"valid_to"`
	ValidStatus string  `json:"valid_status"`
	TaxRate     float64 `json:"tax_rate"`
}

func (TaxDetailsView) TableName() string {
	return "vw_get_tax_setup_details"
}

type TaxSetupBody struct {
	TaxSetup        Tax          `json:"tax_setup"`
	TaxSetupDetails []TaxDetails `json:"tax_setup_details"`
}

type TaxSetupGet struct {
	TaxSetup        []Tax            `json:"tax_setup"`
	TaxSetupDetails []TaxDetailsView `json:"tax_setup_details"`
	TaxSetupView    []TaxView        `json:"tax_setup_view"`
}
