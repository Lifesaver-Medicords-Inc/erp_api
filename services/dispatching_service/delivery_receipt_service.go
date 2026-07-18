package dispatching_services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type DeliveryReceiptService struct {
	CalendarScheduleService          *CalendarScheduleService
	LogisticsCalendarScheduleService *LogisticsCalendarScheduleService
}

// Looks up the sales order's document number via the vw_get_sales_order_dr
// join so it can be shown on the logistics schedule/route without requiring
// the client to look it up separately. Best-effort: returns "" if not found.
func getSalesOrderDocNo(tx *gorm.DB, salesOrderID uint) string {
	var soView models.SalesOrderDrView
	if err := tx.Where("order_id = ?", salesOrderID).First(&soView).Error; err != nil {
		return ""
	}
	return soView.DocumentNo
}

func NewDeliveryReceiptService(calendarScheduleService *CalendarScheduleService, logisticsCalendarScheduleService *LogisticsCalendarScheduleService) *DeliveryReceiptService {
	return &DeliveryReceiptService{
		CalendarScheduleService:          calendarScheduleService,
		LogisticsCalendarScheduleService: logisticsCalendarScheduleService,
	}
}

// Get all delivery receipts with optional conditions
func (s *DeliveryReceiptService) GetDeliveryReceiptsService(conditions map[string]interface{}) ([]dispatching_models.DeliveryReceipt, int, error) {
	var receipts = []dispatching_models.DeliveryReceipt{}

	// Temporarily invalidate cache to force fresh DB fetch
	// key := services.GetKey(&receipts, conditions)
	// services.InvalidateCache(key)

	if err := services.DbGetWithPreloads(&receipts, conditions, "DeliveryReceiptItems", "DeliveryReceiptCosts", "DeliveryReceiptCosts.ReceiptFiles"); err != nil {
		return receipts, fiber.StatusInternalServerError, err
	}

	return receipts, fiber.StatusOK, nil
}

// Get a single delivery receipt
func (s *DeliveryReceiptService) GetDeliveryReceiptService(conditions map[string]interface{}) (*dispatching_models.DeliveryReceipt, int, error) {
	var receipt = &dispatching_models.DeliveryReceipt{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return receipt, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("Order").Preload("ItemReleases").Preload("TripCost")

	for key, val := range conditions {
		query = query.Where(key+" = ?", val)
	}

	// ✅ Already a pointer, so this is fine
	if err := query.First(receipt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.StatusNotFound, err
		}
		return nil, fiber.StatusInternalServerError, err
	}

	return receipt, fiber.StatusOK, nil
}

// Create a new delivery receipt
func (s *DeliveryReceiptService) CreateDeliveryReceiptService(data *dispatching_models.DeliveryReceipt, at models.At) (*dispatching_models.DeliveryReceipt, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return data, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	nextDocNo, err := utils.NextDocNo(tx, new(dispatching_models.DeliveryReceipt), "doc_no")
	if err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	data.DocNo = nextDocNo
	data.ID = 0 // safety: never trust client-provided ID

	if err := services.DbInsert(tx, data); err != nil { // only once
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key") {
			return data, fiber.StatusInternalServerError, errors.New("duplicate record error")
		}
		return data, fiber.StatusInternalServerError, errors.New("failed creating delivery receipt")
	}

	atdata := dispatching_models.DeliveryReceiptAt{
		RefId: data.ID,
		DocNo: strconv.Itoa(data.DocNo),
		At:    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating receiptat")
	}

	schedule := models.CalendarScheduleModel{
		ReferenceDocId: &data.SalesOrderID,
		CalendarScheduleContent: models.CalendarScheduleContent{
			DepartmentType: "Logistics",
			Title:          fmt.Sprintf("Delivery Receipt #%d", data.DocNo),
			StartDate:      data.DeliveryDate,
			EndDate:        data.DeliveryDate,
			Description:    "",
		},
	}

	if _, _, err := s.CalendarScheduleService.CreateCalendarScheduleService(tx, &schedule, at); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating calendar schedule")
	}

	logisticsSchedule := dispatching_models.LogisticsCalendarScheduleModel{
		CalendarScheduleBase: dispatching_models.CalendarScheduleBase{
			Department: "LOGISTICS",
			StartDate:  data.DeliveryDate,
			EndDate:    data.DeliveryDate,
			Title:      fmt.Sprintf("Delivery Receipt #%d", data.DocNo),
		},
		LogisticsCalendarScheduleContent: dispatching_models.LogisticsCalendarScheduleContent{
			SalesOrderId:         data.SalesOrderID,
			SalesOrderDocNo:      getSalesOrderDocNo(tx, data.SalesOrderID),
			DeliveryReceiptId:    data.ID,
			DeliveryReceiptDocNo: strconv.Itoa(data.DocNo),
		},
	}

	if _, _, err := s.LogisticsCalendarScheduleService.CreateLogisticsSchedule(tx, &logisticsSchedule, at); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating logistics calendar schedule")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateDRCaches()

	return data, fiber.StatusCreated, nil
}

// Update an existing delivery receipt
func (s *DeliveryReceiptService) UpdateDeliveryReceiptService(update *dispatching_models.DeliveryReceipt, conditions map[string]interface{}, at models.At) (*dispatching_models.DeliveryReceipt, int, error) {
	var receipt = &dispatching_models.DeliveryReceipt{}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return receipt, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.First(receipt, conditions).Error; err != nil {
		tx.Rollback()
		return nil, fiber.StatusNotFound, err
	}

	if err := services.DbUpdate(tx, update, conditions); err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed updating receipt")
	}

	// ✅ Delete then re-insert items
	if err := tx.Where("delivery_receipt_id = ?", receipt.ID).
		Delete(&dispatching_models.DeliveryReceiptItems{}).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed deleting old items")
	}
	if len(update.DeliveryReceiptItems) > 0 {
		for i := range update.DeliveryReceiptItems {
			update.DeliveryReceiptItems[i].DeliveryReceiptID = receipt.ID
		}
		if err := tx.Create(&update.DeliveryReceiptItems).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed reinserting items")
		}
	}

	// ✅ Delete then re-insert costs
	if err := tx.Where("delivery_receipt_id = ?", receipt.ID).
		Delete(&dispatching_models.DeliveryReceiptCosts{}).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed deleting old costs")
	}
	if len(update.DeliveryReceiptCosts) > 0 {
		for i := range update.DeliveryReceiptCosts {
			update.DeliveryReceiptCosts[i].DeliveryReceiptID = receipt.ID
		}
		if err := tx.Create(&update.DeliveryReceiptCosts).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed reinserting costs")
		}
	}

	// Keep the calendar entries created at DR-create time in sync with the new
	// delivery date (partial update — only touch dates/title, not the rest of
	// the schedule content in case logistics staff edited it separately). If no
	// row exists yet (e.g. this DR predates the calendar-linking feature), create
	// one instead of silently no-oping.
	//
	// The generic tbl_calendar_schedule row only has reference_doc_id (=SalesOrderID),
	// not a delivery_receipt_id, so this match is best-effort if a sales order has
	// multiple delivery receipts.
	genericResult := tx.Model(&models.CalendarScheduleModel{}).
		Where("reference_doc_id = ? AND department_type = ?", receipt.SalesOrderID, "Logistics").
		Updates(map[string]interface{}{
			"start_date": update.DeliveryDate,
			"end_date":   update.DeliveryDate,
			"title":      fmt.Sprintf("Delivery Receipt #%d", receipt.DocNo),
		})
	if genericResult.Error != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed syncing calendar schedule")
	}
	if genericResult.RowsAffected == 0 {
		schedule := models.CalendarScheduleModel{
			ReferenceDocId: &receipt.SalesOrderID,
			CalendarScheduleContent: models.CalendarScheduleContent{
				DepartmentType: "Logistics",
				Title:          fmt.Sprintf("Delivery Receipt #%d", receipt.DocNo),
				StartDate:      update.DeliveryDate,
				EndDate:        update.DeliveryDate,
			},
		}
		if _, _, err := s.CalendarScheduleService.CreateCalendarScheduleService(tx, &schedule, at); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed creating calendar schedule")
		}
	}

	// tbl_dispatching_logistics_calendar_schedule has a real delivery_receipt_id,
	// so this match is reliable.
	logisticsResult := tx.Model(&dispatching_models.LogisticsCalendarScheduleModel{}).
		Where("delivery_receipt_id = ?", receipt.ID).
		Updates(map[string]interface{}{
			"start_date":         update.DeliveryDate,
			"end_date":           update.DeliveryDate,
			"title":              fmt.Sprintf("Delivery Receipt #%d", receipt.DocNo),
			"sales_order_doc_no": getSalesOrderDocNo(tx, receipt.SalesOrderID),
		})
	if logisticsResult.Error != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed syncing logistics calendar schedule")
	}
	if logisticsResult.RowsAffected == 0 {
		logisticsSchedule := dispatching_models.LogisticsCalendarScheduleModel{
			CalendarScheduleBase: dispatching_models.CalendarScheduleBase{
				Department: "LOGISTICS",
				StartDate:  update.DeliveryDate,
				EndDate:    update.DeliveryDate,
				Title:      fmt.Sprintf("Delivery Receipt #%d", receipt.DocNo),
			},
			LogisticsCalendarScheduleContent: dispatching_models.LogisticsCalendarScheduleContent{
				SalesOrderId:         receipt.SalesOrderID,
				SalesOrderDocNo:      getSalesOrderDocNo(tx, receipt.SalesOrderID),
				DeliveryReceiptId:    receipt.ID,
				DeliveryReceiptDocNo: strconv.Itoa(receipt.DocNo),
			},
		}
		if _, _, err := s.LogisticsCalendarScheduleService.CreateLogisticsSchedule(tx, &logisticsSchedule, at); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed creating logistics calendar schedule")
		}
	}

	atdata := dispatching_models.DeliveryReceiptAt{
		RefId: receipt.ID,
		DocNo: strconv.Itoa(receipt.DocNo),
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed creating receiptat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateDRCaches()
	return update, fiber.StatusOK, nil
}

// Delete a delivery receipt
func (s *DeliveryReceiptService) DeleteDeliveryReceiptService(conditions map[string]interface{}, at models.At) (*dispatching_models.DeliveryReceipt, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &dispatching_models.DeliveryReceipt{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	release, status, err := s.GetDeliveryReceiptService(conditions)
	if err != nil {
		return release, status, errors.New("delivery receipt not found")
	}

	if err := services.DbDelete(tx, &release, conditions); err != nil {
		return release, fiber.StatusInternalServerError, errors.New("failed deleting delivery receipt")
	}

	atdata := dispatching_models.DeliveryReceiptAt{RefId: release.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed creating receipt audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateDRCaches()

	return release, fiber.StatusOK, nil
}

// Load SO with approved IR
func (s *DeliveryReceiptService) GetSOWithApprovedIRService(conditions map[string]interface{}) ([]dispatching_models.SalesOrderWithApprovedIRView, int, error) {
	var salesOrder []dispatching_models.SalesOrderWithApprovedIRView

	if err := services.DbGet(&salesOrder, conditions); err != nil {
		return salesOrder, fiber.StatusInternalServerError, errors.New("failed getting so with approved ir")
	}

	return salesOrder, fiber.StatusOK, nil
}

func (s *DeliveryReceiptService) GetSOWithApprovedIRDetailsService(itemReleaseID int64) (interface{}, int, error) {
	var response []dispatching_models.IRDetailsApprovedSOView

	conditions := map[string]interface{}{
		"ItemReleaseId": itemReleaseID,
	}

	if err := services.DbRaw(&response, "sp_GetItemReleaseDetails", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting item release details data")
	}

	return response, fiber.StatusOK, nil
}
func InvalidateDRCaches() {
	cacheKeys := []interface{}{
		dispatching_models.SalesOrderWithApprovedIRView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
