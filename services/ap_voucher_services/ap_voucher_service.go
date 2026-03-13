package ap_voucher_services

import (
	// "errors"

	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type ApVoucherService struct{}

func NewApVoucherService() *ApVoucherService {
	return &ApVoucherService{}
}

func (s *ApVoucherService) GetApVoucher(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.ApVoucherGet

	if err := services.DbGet(&response.ApVoucher, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting ap voucher")
	}

	if err := services.DbGet(&response.ApVoucherDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting ap voucher details")
	}

	return response, fiber.StatusOK, nil
}

func (s *ApVoucherService) GetInvoiceView(conditions map[string]interface{}) (interface{}, int, error) {
	var response []accounting_models.InvoiceReceiptView

	if err := services.DbRaw(&response, "sp_GetInvoiceAPVoucher", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting invoice receipt data")
	}

	return response, fiber.StatusOK, nil
}

func (s *ApVoucherService) CreateApVoucher(body *accounting_models.ApVoucherBody, at models.At) (*accounting_models.ApVoucherBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	nextDocNo, err := utils.NextDocNo(tx, new(accounting_models.ApVoucher), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.ApVoucher.DocNo = nextDocNo

	// Insert main Ap Voucher
	if err := services.DbInsert(tx, &body.ApVoucher); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating ap voucher")
	}

	// Insert Ap Voucher Details
	if err := s.CreateApVoucherDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record
	atdata := accounting_models.ApVoucherAt{
		RefId:            body.ApVoucher.ID,
		ApVoucherContent: body.ApVoucher.ApVoucherContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating ap voucher at")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(accounting_models.InvoiceReceiptView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.APVoucherPaymentView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.APVoucherPaymentDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *ApVoucherService) CreateApVoucherDetails(tx *gorm.DB, body *accounting_models.ApVoucherBody, at models.At) error {
	for i := range body.ApVoucherDetails {
		detail := &body.ApVoucherDetails[i]
		detail.ApVoucherId = body.ApVoucher.ID

		// Insert detail
		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating ap voucher details")
		}

		// Update receipt depending on type
		if err := s.markReceiptAsVouchered(tx, detail.ReceiptType, detail.InvoiceReceiptId); err != nil {
			return err
		}

		// Audit trail
		atdataDetail := accounting_models.ApVoucherDetailsAt{
			RefId:                   detail.ID,
			ApVoucherDetailsContent: detail.ApVoucherDetailsContent,
			At:                      at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating ap voucher details at")
		}
	}

	return nil
}

func (s *ApVoucherService) markReceiptAsVouchered(tx *gorm.DB, receiptType string, receiptId uint) error {
	switch receiptType {
	case "INVOICE RECEIPT":
		res := tx.Model(&accounting_models.InvoiceReceipt{}).
			Where("id = ?", receiptId).
			Update("ap_voucher", true)

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("invoice receipt not found")
		}

	case "BULK INVOICE RECEIPT":
		res := tx.Model(&accounting_models.BulkInvoiceReceipt{}).
			Where("id = ?", receiptId).
			Update("ap_voucher", true)

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("bulk invoice receipt not found")
		}

	default:
		return errors.New("unknown receipt type")
	}

	return nil
}
