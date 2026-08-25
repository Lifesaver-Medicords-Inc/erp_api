package purchase_return_services

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

type PurchaseReturnService struct{}

func NewPurchaseReturnService() *PurchaseReturnService {
	return &PurchaseReturnService{}
}

// GetPurchaseReturn serves both List (conditions == nil) and GetByID
// (conditions == {"ID": x}), same convention as every other Get in this
// codebase.
func (s *PurchaseReturnService) GetPurchaseReturn(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.PurchaseReturnGet

	if err := services.DbGet(&response.PurchaseReturn, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchase return")
	}

	if err := services.DbGet(&response.PurchaseReturnDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchase return details")
	}

	return response, fiber.StatusOK, nil
}

// CreatePurchaseReturn covers the purchaser-initiated path only (§5.8).
// Always created unapproved - see ApprovePurchaseReturn for the gate
// (confirmed with the user: CBDO approves, per §3.2/§16's implication that
// §5.8's own text never spells out).
//
// NOT YET IMPLEMENTED, deliberately: the stock-decrease-on-release effect
// (§5.8) itself, now that the trigger point (approval) is confirmed - and
// auto-generation from a Sales Return line's QtyForPurchaseReturn
// (§12.6.1), the other end of the integration SalesReturnService.
// ApproveSalesReturn's doc comment notes the same gap from the SRT side.
// Both need PRT's UI to exist first to be tested meaningfully.
func (s *PurchaseReturnService) CreatePurchaseReturn(body *models.PurchaseReturnBody, at models.At) (*models.PurchaseReturnBody, int, error) {
	if body.PurchaseReturn.SupplierID == 0 {
		return body, fiber.StatusBadRequest, errors.New("supplier_id is required")
	}
	if body.PurchaseReturn.RefIRID == 0 {
		return body, fiber.StatusBadRequest, errors.New("ref_ir_id is required - a Purchase Return references an Invoice Receipt, not a PO (§5.8)")
	}
	if body.PurchaseReturn.ReturnType != "Return with Debit Memo" && body.PurchaseReturn.ReturnType != "Return without Debit Memo" {
		return body, fiber.StatusBadRequest, errors.New("return_type must be either 'Return with Debit Memo' or 'Return without Debit Memo'")
	}
	for i, d := range body.PurchaseReturnDetails {
		if d.RefIRDetailsID == 0 {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"line %d: ref_ir_details_id is required - match at the IR line level, never at PO level (§5.8)", i+1,
			)
		}
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(models.PurchaseReturn), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}
	body.PurchaseReturn.DocNo = nextDocNo

	// A Purchase Return is always created unapproved - CBDO approval is a
	// deliberate action on an existing draft, never implied by save.
	body.PurchaseReturn.IsApproved = false

	if err := services.DbInsert(tx, &body.PurchaseReturn); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchase return")
	}

	if err := s.createPurchaseReturnDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := models.PurchaseReturnAt{
		RefId:                 body.PurchaseReturn.ID,
		PurchaseReturnContent: body.PurchaseReturn.PurchaseReturnContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchase return at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.PurchaseReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *PurchaseReturnService) createPurchaseReturnDetails(tx *gorm.DB, body *models.PurchaseReturnBody, at models.At) error {
	for i := range body.PurchaseReturnDetails {
		detail := &body.PurchaseReturnDetails[i]
		detail.PurchaseReturnID = body.PurchaseReturn.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating purchase return details")
		}

		atdataDetail := models.PurchaseReturnDetailsAt{
			RefId:                        detail.ID,
			PurchaseReturnDetailsContent: detail.PurchaseReturnDetailsContent,
			At:                           at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating purchase return details at")
		}
	}

	return nil
}

// PurchaseReturnApprovalAccessCode gates approving a Purchase Return -
// confirmed with the user: CBDO, per §3.2's "Purchase Return Approval"
// module-access line and §16's glossary entry, since §5.8's own text never
// spells out an approval step. Same tbl_position_access mechanism as every
// other approval gate in this codebase - grant it to whatever Position
// covers CBDO from the normal Position Access setup screen.
const PurchaseReturnApprovalAccessCode = "PURCHASE_RETURN_APPROVAL"

func (s *PurchaseReturnService) UserCanApprovePurchaseReturn(userId uint) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, PurchaseReturnApprovalAccessCode).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ApprovePurchaseReturn signs off on a pending Purchase Return: permission
// check -> load -> guard it isn't already approved -> stamp who/when ->
// commit. Same shape as SalesReturnService.ApproveSalesReturn.
//
// NOT YET IMPLEMENTED HERE, deliberately: the stock-decrease-on-release
// effect §5.8 describes. This only flips the approval flag and records who
// approved it - see this file's CreatePurchaseReturn doc comment for why
// the effect itself is the next thing to build, not this pass.
func (s *PurchaseReturnService) ApprovePurchaseReturn(purchaseReturnId uint, approvedByUserId uint) (int, error) {
	canApprove, err := s.UserCanApprovePurchaseReturn(approvedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to approve purchase returns")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var purchaseReturn models.PurchaseReturn
	if err := tx.First(&purchaseReturn, purchaseReturnId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("purchase return not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading purchase return")
	}

	if purchaseReturn.IsApproved {
		return fiber.StatusConflict, errors.New("purchase return is already approved")
	}

	var approver models.User
	approverName := ""
	if err := tx.First(&approver, approvedByUserId).Error; err == nil {
		approverName = strings.TrimSpace(approver.FirstName + " " + approver.LastName)
	}

	purchaseReturn.IsApproved = true
	purchaseReturn.ApprovedByID = approvedByUserId
	purchaseReturn.ApprovedByName = approverName
	purchaseReturn.ApprovalDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &purchaseReturn, map[string]interface{}{"id": purchaseReturn.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating purchase return")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.PurchaseReturn{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return fiber.StatusOK, nil
}
