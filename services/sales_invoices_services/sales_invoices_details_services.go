package sales_invoices_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateSalesInvoiceDetails(tx *gorm.DB, childSalesInvoice *[]accounting_models.SalesInvoiceDetail, parentId uint, at models.At) error {

	for i := range *childSalesInvoice {

		child := &(*childSalesInvoice)[i]
		if err := CreateSalesInvoiceDetail(tx, child, parentId, at); err != nil {
			return err
		}
	}
	return nil

}

func CreateSalesInvoiceDetail(tx *gorm.DB, child *accounting_models.SalesInvoiceDetail, parentId uint, at models.At) error {

	child.SalesInvoiceDetailContent.SalesInvoiceId = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to insert sales invoice details")
	}

	childAt := accounting_models.SalesInvoiceDetailAt{
		RefId:                     child.ID,
		SalesInvoiceDetailContent: child.SalesInvoiceDetailContent,
		At:                        at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed to insert sales invoice details at")
	}
	return nil
}
