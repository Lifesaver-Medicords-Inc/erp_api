package debit_memo_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

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

// CreateDebitMemo. §5.19/§12.6.3: commits entirely on SAVE - no draft, no
// approval workflow, ever (§14.57). §14.43: MUST NOT save while
// UNAPPLIED AMOUNT > 0, so every peso of TransAmount has to land on a
// ticked apply row before this succeeds.
//
// NOT YET IMPLEMENTED HERE, deliberately, same reasoning as
// CreditMemoService.CreateCreditMemo: the actual effect on each applied-
// against document's own open balance, and the journal entry write, need
// their own pass against the live accounting code first (CLAUDE.md's
// accounting-inverts-spec-wins rule). This persists the memo and its apply
// table and enforces the document-level rules only.
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
