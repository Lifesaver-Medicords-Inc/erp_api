package debit_memo_services

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

// Fixed debit/credit COA pair, same precedent-following reasoning as
// CreditMemoService's own pair (see its doc comment). A DM lessens what
// SMPC owes: debit Accounts Payable (decreasing the liability), credit the
// closest generic offsetting account in the live chart.
const (
	debitMemoDebitCoaId  uint = 40030 // ACCOUNTS PAYABLE
	debitMemoCreditCoaId uint = 50030 // NON-TRADE EXPENSE
)

// postDebitMemoJournalEntry writes the DM's debit+credit pair (§12.2,
// §12.6.3: "On save... the journal entry is written"). Same period-lookup
// and fixed-pair pattern as postCreditMemoJournalEntry.
func postDebitMemoJournalEntry(tx *gorm.DB, dm *models.DebitMemo, at models.At) error {
	if strings.TrimSpace(dm.DocDate) == "" {
		dm.DocDate = time.Now().Format("01/02/2006")
	}

	var journals []accounting_models.JournalEntry
	if err := tx.Find(&journals).Error; err != nil {
		return errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(dm.DocDate)))
	if err != nil {
		return errors.New("invalid debit memo doc date format")
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
		return errors.New("no active journal entry period is configured for this debit memo's doc date")
	}

	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, debitMemoDebitCoaId).Error; err != nil {
		return errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, debitMemoCreditCoaId).Error; err != nil {
		return errors.New("failed fetching credit chart of account")
	}

	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails {
		entry := accounting_models.JournalEntryDetails{
			JournalEntryDetailsContent: accounting_models.JournalEntryDetailsContent{
				Origin:         "Debit Memo",
				OriginId:       dm.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    dm.DocDate,
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
		createJournalEntry(coaCREDIT, dm.TransAmount, true),
		createJournalEntry(coaDEBIT, dm.TransAmount, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, dm.DocDate, at); err != nil {
			return err
		}
	}

	return nil
}

type DebitMemoService struct{}

func NewDebitMemoService() *DebitMemoService {
	return &DebitMemoService{}
}

// GetDebitMemo serves both List and GetByID, same convention as every
// other Get here.
func (s *DebitMemoService) GetDebitMemo(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.DebitMemoGet

	if err := services.DbGet(&response.DebitMemo, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting debit memo")
	}
	if err := services.DbGet(&response.DebitMemoDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting debit memo details")
	}

	return response, fiber.StatusOK, nil
}

var validTargetDocTypes = map[string]bool{
	"Invoice Receipt":      true,
	"Bulk Invoice Receipt": true,
	"Credit Memo":          true,
}

// applyToTargetDocumentsEpsilon guards the float comparison used to decide
// whether a ticked apply line fully consumed its target (AmountApplied ==
// OpenAmount, allowing for client-entered-money rounding) - same tolerance
// as the §14.43 UNAPPLIED AMOUNT check just above.
const applyToTargetDocumentsEpsilon = 0.005

// applyToTargetDocuments implements the second half of §12.6.3: "A DM's
// save also updates every account it was applied against, not just the
// supplier's." Previously not implemented at all (this DM only recorded its
// own apply-table snapshot of the target, never touched the target itself).
//
// This codebase has no partial running-balance/OPEN AMOUNT tracking for any
// of the three apply-target types (Invoice Receipt, Bulk Invoice Receipt,
// Credit Memo) - confirmed by AP Voucher's own equivalent step
// (markReceiptAsVouchered), which is a boolean "fully vouchered" flag, not a
// numeric balance. So a line that FULLY consumes its target (AmountApplied
// >= OpenAmount) flips that same boolean, consistent with AP Voucher and
// reusing the exact InvoiceReceipt/BulkInvoiceReceipt.ApVoucher field so the
// existing "open only" pickers (sp_GetInvoiceAPVoucher, filtered ap_voucher
// = 0) automatically stop offering it - to either a future AP Voucher or a
// future Debit Memo. Credit Memo gets an equivalent new flag, AppliedByDm,
// since it had no such gate at all.
//
// A PARTIAL application (AmountApplied < OpenAmount) is left as a real,
// flagged gap: there is nowhere to record "this target is 40% consumed"
// without inventing a running-balance column, which is a data-model
// decision, not a bug-fix edit made on my own authority.
func (s *DebitMemoService) applyToTargetDocuments(tx *gorm.DB, body *models.DebitMemoBody) error {
	for _, d := range body.DebitMemoDetails {
		if !d.Apply || d.TargetDocId == 0 {
			continue
		}
		if d.AmountApplied < d.OpenAmount-applyToTargetDocumentsEpsilon {
			continue // partial - no field exists yet to record this, see doc comment above
		}

		switch d.TargetDocType {
		case "Invoice Receipt":
			res := tx.Model(&accounting_models.InvoiceReceipt{}).Where("id = ?", d.TargetDocId).Update("ap_voucher", true)
			if res.Error != nil {
				return res.Error
			}
		case "Bulk Invoice Receipt":
			res := tx.Model(&accounting_models.BulkInvoiceReceipt{}).Where("id = ?", d.TargetDocId).Update("ap_voucher", true)
			if res.Error != nil {
				return res.Error
			}
		case "Credit Memo":
			res := tx.Model(&models.CreditMemo{}).Where("id = ?", d.TargetDocId).Update("applied_by_dm", true)
			if res.Error != nil {
				return res.Error
			}
		}
	}
	return nil
}

// CreateDebitMemo. §5.19/§12.6.3: commits entirely on SAVE - no draft, no
// approval workflow, ever (§14.57). §14.43: MUST NOT save while
// UNAPPLIED AMOUNT > 0, so every peso of TransAmount has to land on a
// ticked apply row before this succeeds.
//
// The journal entry write (§12.2, §12.6.3) is handled below by
// postDebitMemoJournalEntry, following the same accounting-inverts-spec-
// wins diff CreditMemoService's own posting routine did.
//
// §12.6.3's other sentence - "A DM's save also updates every account it was
// applied against" - is handled by applyToTargetDocuments below, for the
// full-consumption case; see that function's doc comment for the partial-
// consumption gap that remains.
func (s *DebitMemoService) CreateDebitMemo(body *models.DebitMemoBody, at models.At) (*models.DebitMemoBody, int, error) {
	dm := &body.DebitMemo

	if dm.SupplierId == 0 {
		return body, fiber.StatusBadRequest, errors.New("supplier_id is required")
	}
	if strings.TrimSpace(dm.ReasonCode) == "" {
		return body, fiber.StatusBadRequest, errors.New("reason_code is required (§14.58)")
	}

	appliedTotal := 0.0
	for i, d := range body.DebitMemoDetails {
		if !d.Apply {
			continue
		}
		if !validTargetDocTypes[d.TargetDocType] {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"line %d: target_doc_type must be 'Invoice Receipt', 'Bulk Invoice Receipt', or 'Credit Memo' - not 'Miscellaneous Receiving' (out of scope, §15)", i+1,
			)
		}
		if d.TargetDocId == 0 {
			return body, fiber.StatusBadRequest, fmt.Errorf("line %d: a ticked apply row requires a target document", i+1)
		}
		appliedTotal += d.AmountApplied
	}

	// §14.43 - MUST NOT save with UNAPPLIED AMOUNT > 0. Comparing via a
	// small epsilon since these are floats coming off client-entered money
	// amounts.
	unapplied := dm.TransAmount - appliedTotal
	if unapplied > 0.005 {
		return body, fiber.StatusBadRequest, fmt.Errorf(
			"unapplied_amount must reach 0 before this can be saved (trans_amount %.2f, applied %.2f, unapplied %.2f) (§14.43)",
			dm.TransAmount, appliedTotal, unapplied,
		)
	}
	dm.UnappliedAmount = 0

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(models.DebitMemo), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}
	dm.DocNo = nextDocNo

	if err := services.DbInsert(tx, dm); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating debit memo")
	}

	if err := s.createDebitMemoDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := s.applyToTargetDocuments(tx, body); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// §12.6.3: committed by SAVE, one step - the journal entry is written
	// right here, same as every other memo.
	if err := postDebitMemoJournalEntry(tx, dm, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := models.DebitMemoAt{
		RefId:            dm.ID,
		DebitMemoContent: dm.DebitMemoContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating debit memo at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(models.DebitMemo{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *DebitMemoService) createDebitMemoDetails(tx *gorm.DB, body *models.DebitMemoBody, at models.At) error {
	for i := range body.DebitMemoDetails {
		detail := &body.DebitMemoDetails[i]
		detail.DebitMemoID = body.DebitMemo.ID
		detail.Balance = detail.OpenAmount - detail.AmountApplied

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating debit memo details")
		}

		atdataDetail := models.DebitMemoDetailsAt{
			RefId:                   detail.ID,
			DebitMemoDetailsContent: detail.DebitMemoDetailsContent,
			At:                      at,
		}
		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating debit memo details at")
		}
	}

	return nil
}
