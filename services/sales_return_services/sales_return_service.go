package sales_return_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type SalesReturnService struct{}

func NewSalesReturnService() *SalesReturnService {
	return &SalesReturnService{}
}

// GetSalesReturn serves both List (conditions == nil) and GetByID
// (conditions == {"ID": x}) the same way every other document's Get
// function in this codebase does - see sales_invoice_services.
func (s *SalesReturnService) GetSalesReturn(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.SalesReturnGet

	if err := services.DbGet(&response.SalesReturn, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales return")
	}

	if err := services.DbGet(&response.SalesReturnDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales return details")
	}

	return response, fiber.StatusOK, nil
}

// CreateSalesReturn. Per §5.13/§12.6.2, saving a Sales Return has NO
// side effects at all - no stock movement, no Purchase Return - it only
// enters the Sales Manager's approval queue. Everything below is just
// persisting the document; see ApproveSalesReturn for the approval-gated
// effects (and its doc comment for what's still NOT implemented there).
func (s *SalesReturnService) CreateSalesReturn(body *models.SalesReturnBody, at models.At) (*models.SalesReturnBody, int, error) {
	// §14.62: REF. DOC. TYPE must be chosen before item selection - at the
	// API boundary that means a create request carrying detail lines but no
	// reference document is rejected outright rather than silently accepted.
	if body.SalesReturn.RefDocType == "" || body.SalesReturn.RefDocID == 0 {
		return body, fiber.StatusBadRequest, errors.New("ref_doc_type and ref_doc_id are required before item selection")
	}
	if body.SalesReturn.RefDocType != "Sales Invoice" && body.SalesReturn.RefDocType != "Delivery Receipt" {
		return body, fiber.StatusBadRequest, errors.New("ref_doc_type must be either 'Sales Invoice' or 'Delivery Receipt'")
	}

	// §14.65: the three-way destination split MUST sum to QTY RECEIVED on
	// every line. The model's own comment defers this check to the service
	// layer (GORM has no cross-column CHECK constraint worth relying on
	// across DB engines) - this is that check.
	for i, d := range body.SalesReturnDetails {
		sum := d.QtyForReplacement + d.QtyToStock + d.QtyForPurchaseReturn
		if sum != d.QtyReceived {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"line %d: qty_for_replacement + qty_to_stock + qty_for_purchase_return (%d) must equal qty_received (%d)",
				i+1, sum, d.QtyReceived,
			)
		}
	}

	// Salesperson/Currency/SalesPeriod/UnitPrice are trusted as sent, not
	// re-derived here - same "freeze by omission" convention already used
	// for tax rate/code everywhere else in this codebase (the client reads
	// the reference document at pick-time and passes its fields straight
	// through; the server never re-joins the live source on write).

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(models.SalesReturn), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}
	body.SalesReturn.DocNo = nextDocNo

	// A Sales Return is always created unapproved - the approval gate is a
	// deliberate action on an existing draft, never implied by save.
	body.SalesReturn.IsApproved = false

	if err := services.DbInsert(tx, &body.SalesReturn); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales return")
	}

	if err := s.createSalesReturnDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := models.SalesReturnAt{
		RefId:              body.SalesReturn.ID,
		SalesReturnContent: body.SalesReturn.SalesReturnContent,
		At:                 at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales return at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.SalesReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *SalesReturnService) createSalesReturnDetails(tx *gorm.DB, body *models.SalesReturnBody, at models.At) error {
	for i := range body.SalesReturnDetails {
		detail := &body.SalesReturnDetails[i]
		detail.SalesReturnID = body.SalesReturn.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating sales return details")
		}

		atdataDetail := models.SalesReturnDetailsAt{
			RefId:                     detail.ID,
			SalesReturnDetailsContent: detail.SalesReturnDetailsContent,
			At:                        at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating sales return details at")
		}
	}

	return nil
}

// SalesReturnApprovalAccessCode is the tbl_position_access code that grants
// a Position the ability to approve a Sales Return - same module-access
// mechanism as ReservationApprovalAccessCode
// (item_stock_services.stock_reservation_service.go). Grant it to the Sales
// Manager position from the normal Position Access setup screen; nothing
// here hardcodes a position name.
const SalesReturnApprovalAccessCode = "SALES_RETURN_APPROVAL"

// UserCanApproveSalesReturn checks whether the given user's Position has
// been granted SalesReturnApprovalAccessCode.
func (s *SalesReturnService) UserCanApproveSalesReturn(userId uint) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, SalesReturnApprovalAccessCode).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ApproveSalesReturn signs off on a pending Sales Return: permission check
// -> load -> guard it isn't already approved -> stamp who/when -> commit.
// Same shape as item_stock_services.setReservationDecision.
//
// NOT YET IMPLEMENTED HERE, deliberately: the two effects §5.13/§12.6.2 say
// approval triggers - stock returning to inventory, and auto-generating a
// Purchase Return for any QtyForPurchaseReturn line - are NOT wired in this
// pass. Both depend on work explicitly scoped to later Phase 1 items (PRT's
// own service layer doesn't exist yet, and the stock-movement call needs to
// go through item_stock_services the same considered way reservations do).
// Wiring those in now, ahead of PRT existing, would mean guessing at an
// interface instead of building against a real one. This function only
// flips the approval flag and records who approved it - the two side
// effects are the very next thing to build once PRT's layer lands.
func (s *SalesReturnService) ApproveSalesReturn(salesReturnId uint, approvedByUserId uint) (int, error) {
	canApprove, err := s.UserCanApproveSalesReturn(approvedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to approve sales returns")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var salesReturn models.SalesReturn
	if err := tx.First(&salesReturn, salesReturnId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("sales return not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading sales return")
	}

	if salesReturn.IsApproved {
		return fiber.StatusConflict, errors.New("sales return is already approved")
	}

	var approver models.User
	approverName := ""
	if err := tx.First(&approver, approvedByUserId).Error; err == nil {
		approverName = strings.TrimSpace(approver.FirstName + " " + approver.LastName)
	}

	salesReturn.IsApproved = true
	salesReturn.ApprovedByID = approvedByUserId
	salesReturn.ApprovedByName = approverName
	salesReturn.ApprovalDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &salesReturn, map[string]interface{}{"id": salesReturn.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating sales return")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.SalesReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return fiber.StatusOK, nil
}
