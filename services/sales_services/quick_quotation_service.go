package sales_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"gorm.io/gorm"
)

// parseQuotationExpiry tries the date formats this codebase actually uses elsewhere
// (they're inconsistent - see invoice_receipt_service.go vs sales_invoice_service.go)
// and returns nil rather than erroring if none match. A nil result just means the
// resulting reservation never gets swept by ExpireStockReservations - it needs manual
// cleanup, which is a known limitation, not a crash.
func parseQuotationExpiry(validUntil string) *time.Time {
	trimmed := strings.TrimSpace(validUntil)
	if trimmed == "" {
		return nil
	}

	for _, layout := range []string{"2006-01-02", "01/02/2006", "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return &t
		}
	}

	return nil
}

// reserveQuotationLineStock places a soft hold for one quotation line - see
// inventory_models.StockReservation for why this doesn't touch physical stock at all.
func reserveQuotationLineStock(tx *gorm.DB, quotationId uint, validUntil string, quick models.SalesQuotationQuick) error {
	if quick.ItemId == 0 || quick.Qty == 0 {
		return nil
	}

	return item_stock_services.NewItemStockService().CreateStockReservation(
		tx, quick.ItemId, quick.Qty, "sales_quotation", quick.ID, quotationId, parseQuotationExpiry(validUntil),
	)
}

// Create Quick Quotation
func CreateSalesQuotationQuick(tx *gorm.DB, parentId uint, QuickQuote models.SalesQuotationQuick, images []models.SalesQuotationSelectedImage, at models.At, validUntil string) error {
	QuickQuote.BasedId = parentId

	if err := services.DbInsert(tx, &QuickQuote); err != nil {
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

	if err := reserveQuotationLineStock(tx, parentId, validUntil, QuickQuote); err != nil {
		return fmt.Errorf("failed reserving stock for quotation line: %w", err)
	}

	if len(images) > 0 {
		if err := CreateSalesQuotationSelectedImages(tx, QuickQuote.ID, images, at); err != nil {
			return err
		}
	}

	return nil
}

// Create Quick Quotation
func CreateSalesQuotationQuickWithSelectedImage(tx *gorm.DB, parentId uint, QuickQuote models.SalesQuotationQuick, at models.At, validUntil string) error {
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

	if err := reserveQuotationLineStock(tx, parentId, validUntil, QuickQuote); err != nil {
		return fmt.Errorf("failed reserving stock for quotation line: %w", err)
	}

	// for _, v := range body.PurchaseOrderDetails {
	// 	conditions := map[string]interface{}{
	// 		"based_id": body.ID,
	// 	}
	// 	if err := UpdatePurchaseOrderDetails(tx, body.ID, v, at, conditions); err != nil {
	// 		return body, fiber.StatusInternalServerError, err
	// 	}
	// }

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

// update quick quotes. validUntil is the parent quotation's ValidUntil - the
// reservation is released and re-placed from scratch with the (possibly changed)
// qty/expiry rather than trying to patch it in place.
func UpdateSalesQuotationQuick(tx *gorm.DB, quickquotes models.SalesQuotationQuick, at models.At, conditions map[string]interface{}, validUntil string) error {
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

	stockService := item_stock_services.NewItemStockService()
	if err := stockService.ReleaseStockReservation(tx, "sales_quotation", quickquotes.ID); err != nil {
		return fmt.Errorf("failed releasing prior stock reservation: %w", err)
	}
	if err := reserveQuotationLineStock(tx, quickquotes.BasedId, validUntil, quickquotes); err != nil {
		return fmt.Errorf("failed re-reserving stock for quotation line: %w", err)
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

	// The quotation line is gone, so whatever it was holding should be too - don't
	// wait for expiry.
	if err := item_stock_services.NewItemStockService().ReleaseStockReservation(tx, "sales_quotation", quickquotes.ID); err != nil {
		return fmt.Errorf("failed releasing stock reservation: %w", err)
	}

	return nil
}
