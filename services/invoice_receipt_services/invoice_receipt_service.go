package invoice_receipt_services

import (
	// "errors"

	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type InvoiceReceiptService struct{}

func NewInvoiceReceiptService() *InvoiceReceiptService {
	return &InvoiceReceiptService{}
}

func (s *InvoiceReceiptService) GetInvoiceReceipt(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.InvoiceReceiptGet

	if err := services.DbGet(&response.InvoiceReceipt, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting invoice receipt")
	}

	if err := services.DbGet(&response.InvoiceReceiptDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting invoice receipt details")
	}

	return response, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) CreateInvoiceReceipt(body *accounting_models.InvoiceReceiptBody, at models.At) (*accounting_models.InvoiceReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	// Insert main Invoice Receipt
	if err := services.DbInsert(tx, &body.InvoiceReceipt); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating invoice receipt")
	}

	body.InvoiceReceipt.DocNo = utils.DocNoGenerator(body.InvoiceReceipt.ID)
	if err := tx.Model(&body.InvoiceReceipt).
		Update("doc_no", body.InvoiceReceipt.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating invoice receipt doc")
	}

	// Insert Invoice Receipt Details
	if err := s.CreateInvoiceReceiptDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Insert audit record for the main request
	atdata := accounting_models.InvoiceReceiptAt{
		RefId:                 body.InvoiceReceipt.ID,
		InvoiceReceiptContent: body.InvoiceReceipt.InvoiceReceiptContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating invoice receipt at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) CreateInvoiceReceiptDetails(tx *gorm.DB, body *accounting_models.InvoiceReceiptBody, at models.At) error {
	for i := range body.InvoiceReceiptDetails {
		detail := &body.InvoiceReceiptDetails[i]
		detail.InvoiceReceiptID = body.InvoiceReceipt.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating invoice receipt details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.InvoiceReceiptDetailsAt{
			RefId:                        detail.ID,
			InvoiceReceiptDetailsContent: detail.InvoiceReceiptDetailsContent,
			At:                           at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating invoice receipt details at")
		}
	}
	return nil
}

func (s *InvoiceReceiptService) UpdateInvoiceReceipt(body *accounting_models.InvoiceReceiptBody, conditions map[string]interface{}, at models.At) (*accounting_models.InvoiceReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	//Update main Invoice Receipt
	if err := services.DbUpdate(tx, &body.InvoiceReceipt, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating invoice receipt")
	}

	// Handle details
	if err := s.UpdateInvoiceReceiptDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.InvoiceReceiptAt{
		RefId:                 body.InvoiceReceipt.ID,
		InvoiceReceiptContent: body.InvoiceReceipt.InvoiceReceiptContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating invoice receipt at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) UpdateInvoiceReceiptDetails(tx *gorm.DB, body *accounting_models.InvoiceReceiptBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.InvoiceReceiptDetails {
		detail := &body.InvoiceReceiptDetails[i]
		detail.InvoiceReceiptID = body.InvoiceReceipt.ID // ensure FK is set

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating invoice receipt details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating invoice receipt details")
			}
		}

		// Audit record for each detail
		atdataDetail := accounting_models.InvoiceReceiptDetailsAt{
			RefId:                        detail.ID,
			InvoiceReceiptDetailsContent: detail.InvoiceReceiptDetailsContent,
			At:                           at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating invoice receipt details at")
		}
	}
	return nil
}

func (s *InvoiceReceiptService) DeleteInvoiceReceipt(body *accounting_models.InvoiceReceiptBody, at models.At) (*accounting_models.InvoiceReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	//Delete main Invoice Receipt
	if err := services.DbDelete(tx, &body.InvoiceReceipt, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting invoice receipt")
	}

	if err := s.DeleteInvoiceReceiptDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.InvoiceReceiptAt{RefId: body.InvoiceReceipt.ID, InvoiceReceiptContent: body.InvoiceReceipt.InvoiceReceiptContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating invoice receipt at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) DeleteInvoiceReceiptDetails(tx *gorm.DB, body *accounting_models.InvoiceReceiptBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &accounting_models.InvoiceReceiptDetails{}, map[string]interface{}{"invoice_receipt_id": body.InvoiceReceipt.ID}); err != nil {
		return errors.New("failed deleting all invoice receipt details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []accounting_models.InvoiceReceiptDetails
	if err := tx.Unscoped().Where("invoice_receipt_id = ?", body.InvoiceReceipt.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := accounting_models.InvoiceReceiptDetailsAt{
				RefId:                        detail.ID,
				InvoiceReceiptDetailsContent: detail.InvoiceReceiptDetailsContent,
				At:                           at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating invoice receipt details audit record")
			}
		}
	}
	return nil
}
