package purchase_return_services

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
//
// NOT YET IMPLEMENTED, deliberately: an approval gate. §5.8's own text never
// describes one, but §3.2 (module access lists "Purchase Return Approval"
// under Admin) and §16's glossary ("CBDO - Executive approver ... purchase
// return") both imply one exists. This is an open question raised back to
// the user rather than guessed at - see the Phase 1 plan's open-questions
// list. Until it's answered, a created Purchase Return here has no
// IsApproved-style field at all (the model doesn't have one), matching the
// literal spec text; the stock-decrease-on-release effect (§5.8) is also not
// wired yet for the same reason - there's no confirmed trigger point to hang
// it on.
//
// Also NOT wired here: auto-generation from a Sales Return line's
// QtyForPurchaseReturn (§12.6.1). That's the other end of the same
// integration this service intentionally leaves open - SalesReturnService.
// ApproveSalesReturn's doc comment notes the same gap from the SRT side.
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
