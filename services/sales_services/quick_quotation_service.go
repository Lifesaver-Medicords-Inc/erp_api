package sales_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Create Quick Quotation
func CreateSalesQuotationQuick(tx *gorm.DB, basedId uint, QuickQuote models.SalesQuotationQuick, at models.At) error {
	content := models.SalesQuotationQuickContent{
		BasedId:         basedId,
		Qty:             QuickQuote.Qty,
		UnitCode:        QuickQuote.UnitCode,
		UnitPrice:       QuickQuote.UnitPrice,
		PercentDiscount: QuickQuote.PercentDiscount,
		NetDiscount:     QuickQuote.NetDiscount,
		NetTotal:        QuickQuote.NetTotal,
		LineTotal:       QuickQuote.LineTotal,
	}

	quickquote := models.SalesQuotationQuick{SalesQuotationQuickContent: content}
	if err := services.DbInsert(tx, &quickquote); err != nil {
		return errors.New("failed creating quick quote")
	}

	quickquotationsat := models.SalesQuotationQuickAt{
		RefId:                      quickquote.ID,
		SalesQuotationQuickContent: content,
		At:                         at,
	}

	if err := services.DbInsert(tx, &quickquotationsat); err != nil {
		return errors.New("failed creating quick quotations")
	}

	return nil
}

// Get Quick quotes
func GetSalesQuotationQuick(quickquotes *models.SalesQuotationQuick, conditions map[string]interface{}) error {
	if err := services.DbGet(quickquotes, conditions); err != nil {
		return errors.New("failed getting quick quotations")
	}

	return nil
}

// update quick quotes
func UpdateItemSpecs(tx *gorm.DB, quickquotes models.SalesQuotationQuick, at models.At) error {
	if err := services.DbUpdate(tx, &quickquotes, nil); err != nil {
		return errors.New("failed updating itemspecs")
	}

	quickquotationsat := models.SalesQuotationQuickAt{
		RefId:                      quickquotes.ID,
		SalesQuotationQuickContent: quickquotes.SalesQuotationQuickContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &quickquotationsat); err != nil {
		return errors.New("failed creating quickquotationsat")
	}

	return nil
}

func DeleteItemSpecs(tx *gorm.DB, quickquotes models.SalesQuotationQuick, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &quickquotes, conditions); err != nil {
		return errors.New("failed deleting itemspecs")
	}

	quickquotationsat := models.SalesQuotationQuickAt{
		RefId:                      quickquotes.ID,
		SalesQuotationQuickContent: quickquotes.SalesQuotationQuickContent,
		At:                         at,
	}
	if err := services.DbInsert(tx, &quickquotationsat); err != nil {
		return errors.New("failed deleting quick quotations at")
	}

	return nil
}
