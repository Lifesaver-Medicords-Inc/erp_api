package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"gorm.io/gorm"
)

// Reservations are no longer placed automatically here - see the RESERVE checkbox on
// StockCheckModal (Quotation.cs) and item_stock_handlers.CreateStockReservation. A sales
// rep/manager has to explicitly check RESERVE for a line; creating or editing a
// quotation line no longer holds stock on its own. validUntil is still threaded through
// Create/Update below only because quotation_service.go already passes it and nothing
// currently uses it - kept rather than ripped out of every call site for what's now a
// no-op.

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

// update quick quotes. If this line already has a manually-placed reservation (see the
// RESERVE checkbox on StockCheckModal), its qty is kept in sync with whatever the user
// just edited QTY to - it does NOT create a reservation for a line that doesn't already
// have one, and editing qty/item here never places a new hold on its own.
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
	if err := stockService.SyncReservationQty(tx, "sales_quotation", quickquotes.ID, quickquotes.Qty); err != nil {
		return fmt.Errorf("failed syncing stock reservation qty: %w", err)
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
	if err := item_stock_services.NewItemStockService().ReleaseStockReservation(tx, "sales_quotation", quickquotes.ID, at.AtUser); err != nil {
		return fmt.Errorf("failed releasing stock reservation: %w", err)
	}

	return nil
}
