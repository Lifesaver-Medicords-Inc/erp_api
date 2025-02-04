package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Create Quick Quotation
func CreateSalesQuotationQuick(tx *gorm.DB, parentId uint, QuickQuote models.SalesQuotationQuick, at models.At) error {

	// parentId = 10015
	QuickQuote.BasedId = parentId

	if err := services.DbInsert(tx, &QuickQuote); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", &QuickQuote)
		return errors.New("failed creating quick quote")
	}

	quickquotationsat := models.SalesQuotationQuickAt{
		RefId:                      QuickQuote.ID,
		SalesQuotationQuickContent: QuickQuote.SalesQuotationQuickContent,
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

// many
func GetSalesQuotationQuicks(quickquotes *[]models.SalesQuotationQuick, conditions map[string]interface{}) error {
	if err := services.DbGet(quickquotes, conditions); err != nil {
		return errors.New("failed getting quick quotations")
	}

	return nil
}

// update quick quotes
func UpdateSalesQuotationQuick(tx *gorm.DB, quickquotes models.SalesQuotationQuick, at models.At, conditions map[string]interface{}) error {

	if err := services.DbUpdate(tx, &quickquotes, conditions); err != nil {
		return errors.New("failed updating quickquotations")
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

func DeleteSalesQuotationQuick(tx *gorm.DB, quickquotes models.SalesQuotationQuick, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &quickquotes, conditions); err != nil {
		return errors.New("failed deleting sales quotation quick")
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
