package job_order_services

import (
	// "errors"

	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type JobOrderService struct{}

func NewJobOrderService() *JobOrderService {
	return &JobOrderService{}
}

func (s *JobOrderService) GetJobOrder(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.JobOrder

	if err := services.DbRaw(&response, "sp_GetJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order data")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetSalesOrderViewEng(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.SalesOrderViewBody

	if err := services.DbGet(&response.SalesOrderView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order")
	}

	if err := services.DbGet(&response.SalesOrderDetailsView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order details")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetEngineerList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.EngineerListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting engineer list")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetComponents(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.Components

	if err := services.DbRaw(&response, "sp_GetComponents", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item components data")
	}

	return response, 0, nil
}

func (s *JobOrderService) CreateJobOrder(body *[]models.JobOrder, at models.At) (*[]models.JobOrder, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	for i := range *body {
		item := &(*body)[i]

		if err := services.DbInsert(tx, item); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return body, fiber.StatusInternalServerError, errors.New("duplicate record error")
			}
			return body, fiber.StatusInternalServerError, errors.New("failed creating job order")
		}

		atdata := models.JobOrderAt{
			RefId:           item.ID,
			JobOrderContent: item.JobOrderContent,
			At:              at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		if err := services.RecomputeSoItemStatus(tx, item.OrderDetailsId); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
		}
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, 0, nil
}

func (s *JobOrderService) UpdateJobOrder(body *[]models.JobOrder, conditions map[string]interface{}, at models.At) (*[]models.JobOrder, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	defer tx.Rollback()

	for i := range *body {
		item := &(*body)[i]

		itemConditions := map[string]interface{}{"id": item.ID}

		// Get the existing job order record
		var existing models.JobOrder
		if err := tx.Where(itemConditions).First(&existing).Error; err != nil {
			return body, fiber.StatusNotFound, fmt.Errorf("job order not found for id %v", item.ID)
		}

		// Handle report_base (replace file if new one is uploaded)
		if item.ReportBase != "" {
			if existing.ReportBase != "" {
				oldPath := "./files/" + existing.ReportBase
				if err := services.DeleteFile(oldPath); err != nil && !os.IsNotExist(err) {
					fmt.Println("Failed deleting old report file:", err)
				}
			}

			filename, err := services.UploadFile(item.ReportBase)
			if err != nil {
				return body, fiber.StatusInternalServerError, fmt.Errorf("failed saving report file for id %v", item.ID)
			}

			item.ReportBase = filename
		} else {
			item.ReportBase = existing.ReportBase
		}

		// Update record in DB
		if err := services.DbUpdate(tx, item, itemConditions); err != nil {
			return body, fiber.StatusInternalServerError, fmt.Errorf("failed updating job order for id %v", item.ID)
		}

		// Insert audit record
		atdata := models.JobOrderAt{
			RefId:           item.ID,
			JobOrderContent: item.JobOrderContent,
			At:              at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		if err := services.RecomputeSoItemStatus(tx, item.OrderDetailsId); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, 0, nil
}

// JobOrderAcceptAccessCode gates "accept SO items for production" (§6.1 (D)) - the
// missing step that splits WAITING ACKNOWLEDGEMENT from WAITING FOR ENGR
// (SO_Item_Status_Module_Spec_2026-08-13.md row 5 vs 6). Grant it to whichever
// Engineering position actually does this from the normal Position Access setup
// screen - nothing here hardcodes a position name.
const JobOrderAcceptAccessCode = "JOB_ORDER_ACCEPT_ACCESS"

// JobOrderWhAckAccessCode gates the warehouse acknowledgement step (§5.23: "acknowledged
// by the Warehouse Manager") that splits CHECKING from IN STOCK (same spec doc, row 9 vs
// 10).
const JobOrderWhAckAccessCode = "JOB_ORDER_WH_ACK_ACCESS"

func userHasAccess(userId uint, code string) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, code).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UserCanAcceptJobOrder checks JobOrderAcceptAccessCode.
func (s *JobOrderService) UserCanAcceptJobOrder(userId uint) (bool, error) {
	return userHasAccess(userId, JobOrderAcceptAccessCode)
}

// UserCanAcknowledgeJobOrder checks JobOrderWhAckAccessCode.
func (s *JobOrderService) UserCanAcknowledgeJobOrder(userId uint) (bool, error) {
	return userHasAccess(userId, JobOrderWhAckAccessCode)
}

// AcceptJobOrder is the missing "accept SO items for production" step (§6.1 (D)) -
// separate from assignment (EngrId/AEngr), so a job can sit accepted-but-unassigned
// (WAITING FOR ENGR) rather than collapsing straight into PENDING/WAITING
// ACKNOWLEDGEMENT the way the UI's existing PENDING-tab filter does today.
func (s *JobOrderService) AcceptJobOrder(jobOrderId uint, acceptedByUserId uint) (int, error) {
	canAccept, err := s.UserCanAcceptJobOrder(acceptedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking accept permission")
	}
	if !canAccept {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to accept job orders")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var jobOrder models.JobOrder
	if err := tx.First(&jobOrder, jobOrderId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("job order not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading job order")
	}

	if jobOrder.IsAccepted {
		return fiber.StatusConflict, errors.New("job order is already accepted")
	}

	var acceptedBy models.User
	acceptedByName := ""
	if err := tx.First(&acceptedBy, acceptedByUserId).Error; err == nil {
		acceptedByName = strings.TrimSpace(acceptedBy.FirstName + " " + acceptedBy.LastName)
	}

	jobOrder.IsAccepted = true
	jobOrder.AcceptedById = acceptedByUserId
	jobOrder.AcceptedByName = acceptedByName
	jobOrder.AcceptedDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &jobOrder, map[string]interface{}{"id": jobOrder.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating job order")
	}

	if err := services.RecomputeSoItemStatus(tx, jobOrder.OrderDetailsId); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// AcknowledgeJobOrder is the missing warehouse-confirmation step (§5.23) that splits
// CHECKING (job Status="COMPLETE" but nobody's confirmed the units landed in stock) from
// IN STOCK (confirmed). Requires the job to already be marked complete - acknowledging a
// job that hasn't finished production yet would be confirming stock that was never
// actually produced.
func (s *JobOrderService) AcknowledgeJobOrder(jobOrderId uint, acknowledgedByUserId uint) (int, error) {
	canAcknowledge, err := s.UserCanAcknowledgeJobOrder(acknowledgedByUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking acknowledgement permission")
	}
	if !canAcknowledge {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to acknowledge job orders (Warehouse Manager only, §5.23)")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var jobOrder models.JobOrder
	if err := tx.First(&jobOrder, jobOrderId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("job order not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading job order")
	}

	if !strings.EqualFold(strings.TrimSpace(jobOrder.Status), "COMPLETE") {
		return fiber.StatusBadRequest, errors.New("this job order is not marked complete yet - nothing to acknowledge")
	}
	if jobOrder.IsWhAcknowledged {
		return fiber.StatusConflict, errors.New("job order is already acknowledged")
	}

	var acknowledgedBy models.User
	acknowledgedByName := ""
	if err := tx.First(&acknowledgedBy, acknowledgedByUserId).Error; err == nil {
		acknowledgedByName = strings.TrimSpace(acknowledgedBy.FirstName + " " + acknowledgedBy.LastName)
	}

	jobOrder.IsWhAcknowledged = true
	jobOrder.WhAcknowledgedById = acknowledgedByUserId
	jobOrder.WhAcknowledgedByName = acknowledgedByName
	jobOrder.WhAcknowledgedDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &jobOrder, map[string]interface{}{"id": jobOrder.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating job order")
	}

	if err := services.RecomputeSoItemStatus(tx, jobOrder.OrderDetailsId); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}
