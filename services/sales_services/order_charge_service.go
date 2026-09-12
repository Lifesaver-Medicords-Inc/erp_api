package sales_services

import (
	"errors"
	"math"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Sales Order cancellation charges (spec 5.4, 8.15).
//
// Two stages, and the split matters: Raise() records the charge and parks the SO
// at FOR REVIEW without touching anything else - stock stays reserved, every
// department's view is unchanged, A/R sees nothing. Only Approve() moves the SO
// to CANCELLED.
//
// CANCELLED is not CLOSED. A cancelled SO carrying a charge still raises a Sales
// Invoice, still enters A/R, and still closes its billing row when the balance
// reaches zero (5.4). Cancellation stops fulfilment, not billing.
type OrderChargeService struct{}

func NewOrderChargeService() *OrderChargeService {
	return &OrderChargeService{}
}

const (
	ChargeStatusForReview = "FOR REVIEW"
	ChargeStatusApproved  = "APPROVED"
	ChargeStatusRejected  = "REJECTED"

	OrderStatusOpen      = "OPEN"
	OrderStatusForReview = "FOR REVIEW"
	OrderStatusCancelled = "CANCELLED"
)

// Computed figures for one cancellation, before anything is stored. The modal
// calls this to show the user what PROCEED would charge.
type ChargePreview struct {
	OrderID                uint    `json:"order_id"`
	DocumentNo             string  `json:"document_no"`
	RestockingFeePercent   float64 `json:"restocking_fee_percent"`
	CancellationFeePercent float64 `json:"cancellation_fee_percent"`
	UndeliveredNet         float64 `json:"undelivered_net"`
	VatRatePercent         float64 `json:"vat_rate_percent"`
	FeeBase                float64 `json:"fee_base"`
	RestockingFee          float64 `json:"restocking_fee"`
	CancellationFee        float64 `json:"cancellation_fee"`
	TotalCharge            float64 `json:"total_charge"`
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

// Compute works out 8.15's figures for one order at the given percentages.
//
//	undelivered net  = Σ (undelivered qty × unit price), less the SO's header
//	                   discounts pro-rated to those lines
//	fee base         = undelivered net × (1 + frozen VAT rate)   <- VAT-INCLUSIVE
//	fee              = fee base × percent
//
// Two things the SO does not store directly have to be derived from what it
// does, and both are deliberate rather than approximations of convenience:
//
//   - The header discounts are not columns on the order. What the order stores is
//     GrossSales (the sum of the lines before header discounts) and
//     TotalAmountDue/VatAmount. TotalAmountDue - VatAmount is exactly 8.2's NET OF
//     VAT, i.e. net sales after the additional discount, so their ratio to
//     GrossSales is the pro-rata factor 8.15 asks for.
//
//   - The frozen VAT rate is not a column either, but VatAmount over NET OF VAT is
//     that same frozen rate (12.1) recovered from the figures the SO froze at
//     creation. Company Setup's current rate is used only when the order carries
//     no VAT figures at all, because today's rate is exactly what 12.1 says not to
//     apply to an old order.
func (s *OrderChargeService) Compute(orderId uint, restockingPercent, cancellationPercent float64) (ChargePreview, int, error) {
	var order models.Order
	if err := initializers.DB.Where("order_id = ?", orderId).Limit(1).Find(&order).Error; err != nil {
		return ChargePreview{}, fiber.StatusInternalServerError, errors.New("failed reading the sales order")
	}
	if order.Order_ID == 0 {
		return ChargePreview{}, fiber.StatusNotFound, errors.New("sales order not found")
	}

	lines := []models.OrderDetails{}
	if err := initializers.DB.Where("based_id = ?", orderId).Find(&lines).Error; err != nil {
		return ChargePreview{}, fiber.StatusInternalServerError, errors.New("failed reading the order lines")
	}

	// What has actually gone out to the customer, per line. Delivery is the
	// Delivery Receipt, not the Item Release: a released unit has left the
	// warehouse but has not necessarily reached the customer, and 8.15 charges on
	// what was not delivered.
	delivered := map[uint]int{}
	receiptItems := []dispatching_models.DeliveryReceiptItems{}
	if err := initializers.DB.
		Joins("JOIN tbl_dispatching_delivery_receipt dr ON dr.id = tbl_dispatching_delivery_receipt_items.delivery_receipt_id").
		Where("dr.sales_order_id = ?", orderId).
		Find(&receiptItems).Error; err == nil {
		for _, item := range receiptItems {
			delivered[item.SalesOrderDetailsId] += item.Qty
		}
	}

	var lineTotal, undeliveredRaw float64
	for _, line := range lines {
		qty := 0
		if line.Qty != nil {
			qty = *line.Qty
		}

		unitPrice := line.ListPrice
		// TotalPrice is what the line actually contributes once its own discount or
		// mark-up is in (8.1); the per-unit price consistent with it is what the fee
		// must use, not the raw list price.
		if qty > 0 && line.TotalPrice > 0 {
			unitPrice = line.TotalPrice / float64(qty)
		}

		lineTotal += unitPrice * float64(qty)

		remaining := qty - delivered[line.OrderDetailsID]
		if remaining <= 0 {
			continue
		}
		undeliveredRaw += unitPrice * float64(remaining)
	}

	// Pro-rate the header discounts onto the undelivered lines.
	netOfVat := order.TotalAmountDue - order.VatAmount
	proRata := 1.0
	if lineTotal > 0 && netOfVat > 0 {
		proRata = netOfVat / lineTotal
		if proRata > 1 {
			proRata = 1
		}
	}
	undeliveredNet := undeliveredRaw * proRata

	// The rate frozen on this order, recovered from its own figures.
	vatRate := 0.0
	if netOfVat > 0 && order.VatAmount > 0 {
		vatRate = order.VatAmount / netOfVat * 100
	} else {
		var company models.CompanyModel
		if err := initializers.DB.Where("id = ?", 1).Limit(1).Find(&company).Error; err == nil {
			vatRate = company.VatRatePercent
		}
	}

	feeBase := undeliveredNet * (1 + vatRate/100)
	restockingFee := feeBase * restockingPercent / 100
	cancellationFee := feeBase * cancellationPercent / 100

	return ChargePreview{
		OrderID:                orderId,
		DocumentNo:             order.Document_No,
		RestockingFeePercent:   restockingPercent,
		CancellationFeePercent: cancellationPercent,
		UndeliveredNet:         round2(undeliveredNet),
		VatRatePercent:         round2(vatRate),
		FeeBase:                round2(feeBase),
		RestockingFee:          round2(restockingFee),
		CancellationFee:        round2(cancellationFee),
		TotalCharge:            round2(restockingFee + cancellationFee),
	}, fiber.StatusOK, nil
}

// Defaults returns the two percentages Company Setup holds (4.5.6), which are what
// the charges modal autofills with. They are defaults only - the modal may send
// anything back, including 0.
func (s *OrderChargeService) Defaults() (float64, float64) {
	var company models.CompanyModel
	if err := initializers.DB.Where("id = ?", 1).Limit(1).Find(&company).Error; err != nil {
		return 0, 0
	}
	return company.RestockingFeePercent, company.CancellationFeePercent
}

// Get returns the charge record in force for an order, if there is one.
func (s *OrderChargeService) Get(orderId uint) (models.SalesOrderCharge, int, error) {
	rows := []models.SalesOrderCharge{}
	if err := initializers.DB.
		Where("order_id = ? AND status <> ?", orderId, ChargeStatusRejected).
		Order("id desc").Limit(1).Find(&rows).Error; err != nil {
		return models.SalesOrderCharge{}, fiber.StatusInternalServerError, errors.New("failed reading the charge record")
	}
	if len(rows) == 0 {
		return models.SalesOrderCharge{}, fiber.StatusNotFound, errors.New("no charge record for this order")
	}
	return rows[0], fiber.StatusOK, nil
}

// Raise is PROCEED on the charges modal: it commits the charge record and puts
// the SO into FOR REVIEW. Nothing else moves until someone approves.
func (s *OrderChargeService) Raise(charge *models.SalesOrderCharge, at models.At) (int, error) {
	if charge.OrderID == 0 {
		return fiber.StatusBadRequest, errors.New("order_id is required")
	}
	if charge.RestockingFeePercent < 0 || charge.CancellationFeePercent < 0 {
		return fiber.StatusBadRequest, errors.New("a fee percentage cannot be negative")
	}

	var order models.Order
	if err := initializers.DB.Where("order_id = ?", charge.OrderID).Limit(1).Find(&order).Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed reading the sales order")
	}
	if order.Order_ID == 0 {
		return fiber.StatusNotFound, errors.New("sales order not found")
	}
	if strings.EqualFold(order.Status, OrderStatusCancelled) {
		return fiber.StatusBadRequest, errors.New("this order is already cancelled")
	}

	// One live cancellation at a time.
	existing := []models.SalesOrderCharge{}
	if err := initializers.DB.Where("order_id = ? AND status = ?", charge.OrderID, ChargeStatusForReview).
		Limit(1).Find(&existing).Error; err == nil && len(existing) > 0 {
		return fiber.StatusBadRequest, errors.New("a cancellation is already awaiting approval on this order")
	}

	// The figures are recomputed here from the percentages the user sent rather
	// than trusted off the request - the modal is allowed to set the rates, not
	// the arithmetic.
	preview, status, err := s.Compute(charge.OrderID, charge.RestockingFeePercent, charge.CancellationFeePercent)
	if err != nil {
		return status, err
	}

	charge.DocumentNo = preview.DocumentNo
	charge.UndeliveredNet = preview.UndeliveredNet
	charge.VatRatePercent = preview.VatRatePercent
	charge.FeeBase = preview.FeeBase
	charge.RestockingFee = preview.RestockingFee
	charge.CancellationFee = preview.CancellationFee
	charge.TotalCharge = preview.TotalCharge
	charge.Status = ChargeStatusForReview

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Create(charge).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed creating the charge record")
	}

	// The SO parks at FOR REVIEW. Deliberately the only thing touched: reserved
	// stock, the departments' views and A/R all stay exactly as they were (5.4).
	if err := tx.Model(&models.Order{}).Where("order_id = ?", charge.OrderID).
		UpdateColumns(map[string]interface{}{"status": OrderStatusForReview}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed putting the order under review")
	}

	if err := s.audit(tx, charge, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

// Approve is the Sales Manager's or CBDO's decision. Only here does the SO become
// CANCELLED, and only here does the charge record become something A/R can
// invoice.
//
// Releasing the reserved stock is NOT done here. Reservations are the warehouse's
// own submodule with its own expiry sweep (10.4), and having two places write
// that table is how a release goes missing; the cancellation makes the order
// CANCELLED and the reservation path reads that.
func (s *OrderChargeService) Approve(id uint, approve bool, reviewedBy string, reviewedById uint, reviewedDate string, at models.At) (int, error) {
	rows := []models.SalesOrderCharge{}
	if err := initializers.DB.Where("id = ?", id).Limit(1).Find(&rows).Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed reading the charge record")
	}
	if len(rows) == 0 {
		return fiber.StatusNotFound, errors.New("charge record not found")
	}

	charge := rows[0]
	if charge.Status != ChargeStatusForReview {
		return fiber.StatusBadRequest, errors.New("this cancellation has already been decided")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	chargeStatus := ChargeStatusRejected
	// Rejection returns the SO to OPEN, unchanged and with no charge in force
	// (5.4 step 5). The rejected row is kept rather than deleted - it is the record
	// that someone asked.
	orderStatus := OrderStatusOpen
	if approve {
		chargeStatus = ChargeStatusApproved
		orderStatus = OrderStatusCancelled
	}

	if err := tx.Model(&models.SalesOrderCharge{}).Where("id = ?", id).
		UpdateColumns(map[string]interface{}{
			"status":         chargeStatus,
			"reviewed_by":    reviewedBy,
			"reviewed_by_id": reviewedById,
			"reviewed_date":  reviewedDate,
		}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed recording the decision")
	}

	if err := tx.Model(&models.Order{}).Where("order_id = ?", charge.OrderID).
		UpdateColumns(map[string]interface{}{"status": orderStatus}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed updating the order status")
	}

	charge.Status = chargeStatus
	charge.ReviewedBy = reviewedBy
	charge.ReviewedById = reviewedById
	charge.ReviewedDate = reviewedDate
	if err := s.audit(tx, &charge, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

// Pending lists the cancellations awaiting a decision, for the approver's queue.
func (s *OrderChargeService) Pending() ([]models.SalesOrderCharge, int, error) {
	rows := []models.SalesOrderCharge{}
	if err := initializers.DB.Where("status = ?", ChargeStatusForReview).
		Order("id desc").Find(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed reading pending cancellations")
	}
	return rows, fiber.StatusOK, nil
}

func (s *OrderChargeService) audit(tx *gorm.DB, charge *models.SalesOrderCharge, at models.At) error {
	audit := models.SalesOrderChargeAt{
		RefId:                   charge.ID,
		SalesOrderChargeContent: charge.SalesOrderChargeContent,
		At:                      at,
	}
	if err := services.DbInsert(tx, &audit); err != nil {
		return errors.New("failed writing the charge audit row")
	}
	return nil
}
