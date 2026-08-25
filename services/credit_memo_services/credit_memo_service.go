package credit_memo_services

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

type CreditMemoService struct{}

func NewCreditMemoService() *CreditMemoService {
	return &CreditMemoService{}
}

// GetCreditMemo serves both List (conditions == nil) and GetByID
// (conditions == {"ID": x}), same convention as every other Get here.
func (s *CreditMemoService) GetCreditMemo(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.CreditMemoGet

	if err := services.DbGet(&response.CreditMemo, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting credit memo")
	}

	return response, fiber.StatusOK, nil
}

// partnerHasEntityType checks tbl_bpi_entity for whether the given BPI is
// registered under the given tbl_setup_bpi_entity code ("CUS" or "SUP") -
// a partner can legitimately hold both at once, so this checks membership,
// not a single resolved "the" class.
func partnerHasEntityType(bpiGeneralId uint, entityCode string) (bool, error) {
	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_bpi_entity be
		INNER JOIN tbl_setup_bpi_entity e ON e.id = be.entity_id
		WHERE be.bpi_general_id = ? AND e.code = ?
	`, bpiGeneralId, entityCode).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCreditMemo. §5.18/§14.98: direction is derived from PartnerType and
// is NEVER a free client choice - a control that lets someone pick is the
// one way this document posts a customer credit as a payable. In practice
// that means: body.CreditMemo.PartnerType is fixed by which module/screen
// is calling (the A/P Credit Memo screen always sends "Supplier"; A/R's own
// path always sends "Customer" - the client never renders it as an editable
// dropdown), and this function additionally verifies server-side that the
// partner actually holds that BPI entity type (a partner can be both a
// customer and a supplier at once, so this is a membership check, not a
// single ambiguous lookup).
//
// NOT YET IMPLEMENTED HERE, deliberately: the actual balance/ledger effect
// (adding to the supplier's payable, reducing the customer's receivable)
// and the journal entry write (§12.2, §12.6.3). CLAUDE.md's own rule for
// this codebase is that the accounting module inverts "the spec wins" -
// production ledger-posting code may already differ from the spec on
// purpose, so wiring a posting routine here without first diffing against
// the LIVE accounting code (not just this spec section) risks reintroducing
// exactly the failure §12.6.2 warns about ("the single most damaging
// failure this document has"). That check is its own pass. This function
// persists the document and enforces the document-level rules (direction,
// required reason code, the customer-side approval gate, and §14.100's
// supplier-only fields) - nothing more.
func (s *CreditMemoService) CreateCreditMemo(body *models.CreditMemoBody, at models.At) (*models.CreditMemoBody, int, error) {
	cm := &body.CreditMemo

	if cm.PartnerId == 0 {
		return body, fiber.StatusBadRequest, errors.New("partner_id is required")
	}
	if cm.PartnerType != "Supplier" && cm.PartnerType != "Customer" {
		return body, fiber.StatusBadRequest, errors.New("partner_type must be either 'Supplier' or 'Customer'")
	}
	if strings.TrimSpace(cm.ReasonCode) == "" {
		return body, fiber.StatusBadRequest, errors.New("reason_code is required (§14.58)")
	}

	entityCode := "SUP"
	if cm.PartnerType == "Customer" {
		entityCode = "CUS"
	}
	hasType, err := partnerHasEntityType(cm.PartnerId, entityCode)
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed checking partner entity type")
	}
	if !hasType {
		return body, fiber.StatusBadRequest, fmt.Errorf("partner %d is not registered as a %s", cm.PartnerId, cm.PartnerType)
	}

	// §14.100: a customer Credit Memo MUST NOT offer DM REFUND / REF. DM
	// NO. - those are supplier-only. Reject rather than silently drop, so a
	// caller sending them finds out instead of losing data quietly.
	if cm.PartnerType == "Customer" {
		if (cm.DmRefund != nil && *cm.DmRefund) || cm.RefDmNo != "" || cm.RefDmId != 0 {
			return body, fiber.StatusBadRequest, errors.New("a customer Credit Memo cannot carry DM REFUND / REF. DM NO. - those are supplier-only (§14.100)")
		}
	}

	// Approval state on create is always the zero value, for both sides:
	// customer starts unapproved and stays that way until ApproveCreditMemo
	// runs (§14.99 - SAVE must never move it there itself); supplier never
	// uses this gate at all (§12.6.3), so it's simply unused for that row.
	cm.IsApproved = false
	cm.ApprovedByID = 0
	cm.ApprovedByName = ""
	cm.ApprovalDate = ""

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(models.CreditMemo), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}
	cm.DocNo = nextDocNo

	if err := services.DbInsert(tx, cm); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating credit memo")
	}

	atdata := models.CreditMemoAt{
		RefId:             cm.ID,
		CreditMemoContent: cm.CreditMemoContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating credit memo at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.CreditMemo{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

// CreditMemoApprovalAccessCode gates approving a CUSTOMER Credit Memo -
// §3.3: COO only. Same tbl_position_access mechanism as every other
// approval gate in this codebase.
const CreditMemoApprovalAccessCode = "CREDIT_MEMO_APPROVAL"

func (s *CreditMemoService) UserCanApproveCreditMemo(userId uint) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, CreditMemoApprovalAccessCode).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ApproveCreditMemo. §14.99: only a customer Credit Memo has an approval
// step at all, and only the COO may perform it - nothing reaches the
// receivable until this runs. A supplier Credit Memo is rejected outright
// (§14.57/§12.6.3 - it already committed on save, there's nothing to
// approve).
//
// NOT YET IMPLEMENTED HERE, deliberately, same reasoning as Create: the
// actual receivable-reduction and journal entry write (§6.3 - "this is the
// only event that credits a customer") needs its own pass against the live
// accounting code first. This flips the approval flag and records who/when.
func (s *CreditMemoService) ApproveCreditMemo(creditMemoId uint, approvedByUserId uint) (int, error) {
	canApprove, err := s.UserCanApproveCreditMemo(approvedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user is not authorized to approve credit memos (COO only)")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var cm models.CreditMemo
	if err := tx.First(&cm, creditMemoId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("credit memo not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading credit memo")
	}

	if cm.PartnerType != "Customer" {
		return fiber.StatusBadRequest, errors.New("only a customer credit memo requires approval - a supplier credit memo commits on save (§12.6.3)")
	}
	if cm.IsApproved {
		return fiber.StatusConflict, errors.New("credit memo is already approved")
	}

	var approver models.User
	approverName := ""
	if err := tx.First(&approver, approvedByUserId).Error; err == nil {
		approverName = strings.TrimSpace(approver.FirstName + " " + approver.LastName)
	}

	cm.IsApproved = true
	cm.ApprovedByID = approvedByUserId
	cm.ApprovedByName = approverName
	cm.ApprovalDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &cm, map[string]interface{}{"id": cm.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating credit memo")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.CreditMemo{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return fiber.StatusOK, nil
}
