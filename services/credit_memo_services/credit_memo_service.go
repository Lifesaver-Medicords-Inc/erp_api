package credit_memo_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/journal_entry_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

// Fixed debit/credit COA pair per side, matching the precedent every other
// auto-posting document in this codebase already follows (Sales Invoice,
// Invoice Receipt, Payment Voucher, Payment Receipt all hardcode a fixed
// pair rather than let the user pick one - §12.2 says otherwise, but the
// accounting module inverts "the spec wins": production's own code is what
// actually ships, so this follows it rather than building a GL-account
// picker nothing else in the system has). Closest fit in the live chart -
// there's no dedicated Purchase Discount / Sales Returns contra-account yet.
const (
	creditMemoSupplierDebitCoaId  uint = 50030 // NON-TRADE EXPENSE
	creditMemoSupplierCreditCoaId uint = 40030 // ACCOUNTS PAYABLE
	creditMemoCustomerDebitCoaId  uint = 70037 // SALES
	creditMemoCustomerCreditCoaId uint = 70032 // TRADE RECEIVABLE
)

// postCreditMemoJournalEntry writes the CM's debit+credit pair (§12.2).
// Supplier side: adds a payable => debit an expense, credit Accounts
// Payable. Customer side: reduces a receivable => debit Sales (reversing
// the recognized revenue), credit Trade Receivable. Called from
// CreateCreditMemo for a supplier CM (commits on SAVE, §12.6.3) and from
// ApproveCreditMemo for a customer CM (nothing posts until the COO
// approves it - §5.18).
func postCreditMemoJournalEntry(tx *gorm.DB, cm *models.CreditMemo, at models.At) error {
	if strings.TrimSpace(cm.DocDate) == "" {
		cm.DocDate = time.Now().Format("01/02/2006")
	}

	var journals []accounting_models.JournalEntry
	if err := tx.Find(&journals).Error; err != nil {
		return errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(cm.DocDate)))
	if err != nil {
		return errors.New("invalid credit memo doc date format")
	}

	var matchedJournalID uint
	for _, j := range journals {
		parts := strings.Split(j.Period, " to ")
		if len(parts) != 2 {
			continue
		}
		startDate, err1 := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(parts[0]))
		endDate, err2 := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		if !docDate.Before(startDate) && !docDate.After(endDate) {
			matchedJournalID = j.ID
			break
		}
	}
	if matchedJournalID == 0 {
		return errors.New("no active journal entry period is configured for this credit memo's doc date")
	}

	debitCoaId, creditCoaId := creditMemoSupplierDebitCoaId, creditMemoSupplierCreditCoaId
	if cm.PartnerType == "Customer" {
		debitCoaId, creditCoaId = creditMemoCustomerDebitCoaId, creditMemoCustomerCreditCoaId
	}

	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, debitCoaId).Error; err != nil {
		return errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, creditCoaId).Error; err != nil {
		return errors.New("failed fetching credit chart of account")
	}

	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails {
		entry := accounting_models.JournalEntryDetails{
			JournalEntryDetailsContent: accounting_models.JournalEntryDetailsContent{
				Origin:         "Credit Memo",
				OriginId:       cm.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    cm.DocDate,
				CreatedBy:      at.AtUser,
				AccountTitle:   account.Name,
				PostingRef:     account.Code,
				PostingRefId:   account.ID,
				JournalEntryId: matchedJournalID,
			},
		}
		if isCredit {
			entry.Credit = amount
		} else {
			entry.Debit = amount
		}
		return entry
	}

	jeService := journal_entry_services.NewJournalEntryService2()

	for _, e := range []accounting_models.JournalEntryDetails{
		createJournalEntry(coaCREDIT, cm.TransAmount, true),
		createJournalEntry(coaDEBIT, cm.TransAmount, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, cm.DocDate, at); err != nil {
			return err
		}
	}

	return nil
}

type CreditMemoService struct{}

func NewCreditMemoService() *CreditMemoService {
	return &CreditMemoService{}
}

// GetCreditMemo serves both List (conditions == nil) and GetByID
// (conditions == {"ID": x}), same convention as every other Get here.
//
// Fills in OpenAmount for every supplier CM in the result - a live query per
// row rather than a join, but this list is never large enough (one company's
// supplier credit memos) for that to matter, and it keeps
// ComputeCreditMemoOpenAmount as the single source of truth instead of
// duplicating its SQL into this query too.
func (s *CreditMemoService) GetCreditMemo(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.CreditMemoGet

	if err := services.DbGet(&response.CreditMemo, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting credit memo")
	}

	for i := range response.CreditMemo {
		cm := &response.CreditMemo[i]
		if cm.PartnerType != "Supplier" {
			continue
		}
		openAmount, err := services.ComputeCreditMemoOpenAmount(initializers.DB, cm.ID)
		if err != nil {
			continue // don't fail the whole list over one row's computation
		}
		cm.OpenAmount = openAmount
	}

	return response, fiber.StatusOK, nil
}

// partnerHasEntityType checks tbl_bpi_entity for whether the given BPI is
// registered under the given tbl_setup_bpi_entity code ("CUS" or "SUP") -
// a partner can legitimately hold both at once, so this checks membership,
// not a single resolved "the" class.
// GetCreditMemoCustomers backs the Customer Credit Memo screen's own partner
// picker (vw_get_credit_memo_customer). Deliberately NOT vw_get_customer /
// SalesInvoiceService.GetCustomer: that one exposes the parent tbl_bpi.id as
// customer_id, which partnerHasEntityType below can never match (tbl_bpi_entity
// keys on bpi_general_id), so every customer CM raised through the Sales
// Invoice picker failed its own registration guard. This view returns the
// branch id and is already filtered to partners actually holding "CUS", so the
// picker can't offer one that would fail.
func (s *CreditMemoService) GetCreditMemoCustomers() (interface{}, int, error) {
	var response []accounting_models.CreditMemoCustomerView

	if err := services.DbGet(&response, nil); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting credit memo customers")
	}

	return response, fiber.StatusOK, nil
}

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
// The journal entry write (§12.2, §12.6.3) is handled by
// postCreditMemoJournalEntry, called below for a supplier CM (posts on this
// same SAVE) - a customer CM posts nothing here; see ApproveCreditMemo.
// That posting routine was deliberately NOT built until the live
// SI/IR/PV/PAYR posting code had been diffed against §12.2 first (the
// accounting module inverts "the spec wins" in this codebase) - wiring one
// blind risked reintroducing exactly the failure §12.6.2 warns about ("the
// single most damaging failure this document has": posting a customer
// credit as a payable). That diff is done; see postCreditMemoJournalEntry's
// own doc comment for the resulting fixed debit/credit pair.
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

	// §12.6.3: a supplier CM commits - and posts - in one step, on SAVE. A
	// customer CM posts nothing here; ApproveCreditMemo does it, and only
	// once the COO approves (§5.18) - "nothing reaches the receivable until
	// then".
	if cm.PartnerType == "Supplier" {
		if err := postCreditMemoJournalEntry(tx, cm, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
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
func (s *CreditMemoService) ApproveCreditMemo(creditMemoId uint, approvedByUserId uint, at models.At) (int, error) {
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

	// Only now does the receivable actually move (§5.18, §12.6.2 step 4:
	// "Only that approval moves the receivable and writes the journal
	// entry").
	if err := postCreditMemoJournalEntry(tx, &cm, at); err != nil {
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.CreditMemo{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return fiber.StatusOK, nil
}
