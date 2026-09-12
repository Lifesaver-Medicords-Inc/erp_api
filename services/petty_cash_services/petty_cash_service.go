package petty_cash_services

import (
	"errors"
	"math"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Petty Cash Replenishment (spec 5.26).
//
// The figures are recomputed here from the rows rather than trusted from the
// client, and the tally check is enforced on submit and on approval: a request
// whose CASH ON HAND + FOR REIMBURSEMENT does not equal the accountable fund
// cannot be submitted (5.26, exact allocation per 2.9).
type PettyCashService struct{}

func NewPettyCashService() *PettyCashService {
	return &PettyCashService{}
}

const (
	StatusDraft     = "DRAFT"
	StatusSubmitted = "SUBMITTED"
	StatusApproved  = "APPROVED"
)

// Money either side of a peso-centavo rounding difference is equal.
const tolerance = 0.005

func (s *PettyCashService) GetAll() ([]accounting_models.PettyCash, int, error) {
	rows := []accounting_models.PettyCash{}
	if err := initializers.DB.Preload("Details").Order("doc_no desc").Find(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed reading petty cash requests")
	}
	return rows, fiber.StatusOK, nil
}

func (s *PettyCashService) Get(id uint) (accounting_models.PettyCash, int, error) {
	rows := []accounting_models.PettyCash{}
	if err := initializers.DB.Preload("Details").Where("id = ?", id).Limit(1).Find(&rows).Error; err != nil {
		return accounting_models.PettyCash{}, fiber.StatusInternalServerError, errors.New("failed reading the petty cash request")
	}
	if len(rows) == 0 {
		return accounting_models.PettyCash{}, fiber.StatusNotFound, errors.New("petty cash request not found")
	}
	return rows[0], fiber.StatusOK, nil
}

// Compute fills the derived figures from the rows. TOTAL EXPENSE is the sum of
// the ordinary rows; ENCASHMENT is the offsetting negative line; FOR
// REIMBURSEMENT is their sum; ACCOUNTABILITY adds the counted cash on hand.
func Compute(request *accounting_models.PettyCash) {
	var expense, encashment float64
	for _, d := range request.Details {
		if d.IsEncashment {
			encashment += d.Amount
			continue
		}
		expense += d.Amount
	}

	request.TotalExpense = expense
	request.Encashment = encashment
	request.ForReimbursement = expense + encashment + request.Deductions
	request.Accountability = request.ForReimbursement + request.CashOnHand
}

// Tallies reports whether CASH ON HAND + FOR REIMBURSEMENT equals the fund the
// custodian is accountable for.
func Tallies(request accounting_models.PettyCash) bool {
	return math.Abs(request.CashOnHand+request.ForReimbursement-request.AccountableFund) <= tolerance
}

func (s *PettyCashService) Save(request *accounting_models.PettyCash, at models.At) (int, error) {
	if strings.TrimSpace(request.CutOffDate) == "" {
		return fiber.StatusBadRequest, errors.New("the request needs a cut-off date")
	}

	Compute(request)

	if request.Status == "" {
		request.Status = StatusDraft
	}
	if request.Status != StatusDraft && !Tallies(*request) {
		return fiber.StatusBadRequest, errors.New(
			"cash on hand plus for reimbursement must equal the accountable fund before this request can be submitted (spec 5.26)")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	isNew := request.ID == 0
	if isNew {
		var highest []accounting_models.PettyCash
		if err := tx.Order("doc_no desc").Limit(1).Find(&highest).Error; err == nil && len(highest) > 0 {
			request.DocNo = highest[0].DocNo + 1
		} else {
			request.DocNo = 1
		}

		if err := tx.Omit("Details").Create(request).Error; err != nil {
			tx.Rollback()
			return fiber.StatusInternalServerError, errors.New("failed creating the petty cash request")
		}
	} else {
		if err := tx.Model(&accounting_models.PettyCash{}).Where("id = ?", request.ID).
			Select("*").Omit("id", "Details").Updates(request).Error; err != nil {
			tx.Rollback()
			return fiber.StatusInternalServerError, errors.New("failed updating the petty cash request")
		}
		if err := tx.Where("petty_cash_id = ?", request.ID).
			Delete(&accounting_models.PettyCashDetails{}).Error; err != nil {
			tx.Rollback()
			return fiber.StatusInternalServerError, errors.New("failed replacing the petty cash rows")
		}
	}

	for i := range request.Details {
		request.Details[i].ID = 0
		request.Details[i].PettyCashId = request.ID
		if err := tx.Create(&request.Details[i]).Error; err != nil {
			tx.Rollback()
			return fiber.StatusInternalServerError, errors.New("failed saving a petty cash row")
		}
		detailAudit := accounting_models.PettyCashDetailsAt{
			RefId:                   request.Details[i].ID,
			PettyCashDetailsContent: request.Details[i].PettyCashDetailsContent,
			At:                      at,
		}
		if err := services.DbInsert(tx, &detailAudit); err != nil {
			tx.Rollback()
			return fiber.StatusInternalServerError, errors.New("failed writing a petty cash row audit")
		}
	}

	if err := s.audit(tx, request, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

// Approve is the COO's step: the fund is replenished back to its ceiling and the
// custodian adds the receipt back to the cash on hand.
func (s *PettyCashService) Approve(id uint, approvedBy string, approvedDate string, at models.At) (int, error) {
	request, status, err := s.Get(id)
	if err != nil {
		return status, err
	}
	if request.Status == StatusApproved {
		return fiber.StatusBadRequest, errors.New("this request is already approved")
	}

	Compute(&request)
	if !Tallies(request) {
		return fiber.StatusBadRequest, errors.New(
			"cash on hand plus for reimbursement must equal the accountable fund before this request can be approved (spec 5.26)")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Model(&accounting_models.PettyCash{}).Where("id = ?", id).
		UpdateColumns(map[string]interface{}{
			"status":        StatusApproved,
			"approved_by":   approvedBy,
			"approved_date": approvedDate,
		}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed approving the petty cash request")
	}

	request.Status = StatusApproved
	request.ApprovedBy = approvedBy
	request.ApprovedDate = approvedDate
	if err := s.audit(tx, &request, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

// NextCycle opens the next request as a copy of the latest one: rows cleared, and
// the balance carried forward explicitly (5.26). Nothing is saved until the
// custodian saves it.
func (s *PettyCashService) NextCycle() (accounting_models.PettyCash, int, error) {
	rows := []accounting_models.PettyCash{}
	if err := initializers.DB.Preload("Details").Order("doc_no desc").Limit(1).Find(&rows).Error; err != nil {
		return accounting_models.PettyCash{}, fiber.StatusInternalServerError, errors.New("failed reading the last cycle")
	}

	next := accounting_models.PettyCash{}
	next.Status = StatusDraft
	next.Details = []accounting_models.PettyCashDetails{}

	if len(rows) > 0 {
		last := rows[0]
		Compute(&last)
		next.AccountableFund = last.AccountableFund
		// The fund is replenished to its ceiling, so the next cycle opens on it.
		next.OpeningFund = last.AccountableFund
	}

	return next, fiber.StatusOK, nil
}

func (s *PettyCashService) Delete(id uint, at models.At) (int, error) {
	request, status, err := s.Get(id)
	if err != nil {
		return status, err
	}
	if request.Status == StatusApproved {
		return fiber.StatusBadRequest, errors.New("an approved request cannot be deleted")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where("petty_cash_id = ?", id).Delete(&accounting_models.PettyCashDetails{}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed deleting the petty cash rows")
	}
	if err := services.DbDelete(tx, &accounting_models.PettyCash{}, map[string]interface{}{"id": id}); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed deleting the petty cash request")
	}
	if err := s.audit(tx, &request, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

func (s *PettyCashService) audit(tx *gorm.DB, request *accounting_models.PettyCash, at models.At) error {
	audit := accounting_models.PettyCashAt{
		RefId:            request.ID,
		PettyCashContent: request.PettyCashContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &audit); err != nil {
		return errors.New("failed writing the petty cash audit row")
	}
	return nil
}
