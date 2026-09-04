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

	// "Order", "ItemReleases" and "TripCost" are not associations on
	// DeliveryReceiptContent - it declares DeliveryReceiptItems and
	// DeliveryReceiptCosts and nothing else. GORM rejected the whole query with
	// "unsupported relations", so this getter ALWAYS failed: GET /delivery-receipt/:id
	// returned not-found for every id, and DeleteDeliveryReceiptService (which calls
	// this first) could never delete anything.
	query := tx.
		Preload("DeliveryReceiptItems").
		Preload("DeliveryReceiptCosts").
		Preload("DeliveryReceiptCosts.ReceiptFiles")

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

	// Same safety, extended to the nested items - this field wasn't covered by the
	// header-only reset above, and a non-zero client-supplied id here would make
	// GORM's association cascade upsert (MATCH-if-found, INSERT-if-not) instead of a
	// plain insert of new rows, which is both a correctness risk (a malformed/stale
	// payload could silently overwrite an unrelated existing row by id) and - the way
	// this was actually found - the exact shape SQL Server refuses outright once the
	// target table has any trigger (see RecomputeSoItemStatus's own doc comment).
	for i := range data.DeliveryReceiptItems {
		data.DeliveryReceiptItems[i].ID = 0
	}

	// Cost rows are shared with the logistics route now (§13.3), so both halves of
	// COST TYPE and the computed TOTAL COST are filled in before the insert cascades
	// them - see NormalizeDeliveryReceiptCosts.
	if err := NormalizeDeliveryReceiptCosts(tx, data.DeliveryReceiptCosts); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed preparing delivery costs")
	}

	if err := services.DbInsert(tx, data); err != nil { // only once
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key") {
			return data, fiber.StatusInternalServerError, errors.New("duplicate record error")
		}
		return data, fiber.StatusInternalServerError, errors.New("failed creating delivery receipt")
	}

	for _, item := range data.DeliveryReceiptItems {
		if err := services.RecomputeSoItemStatus(tx, item.SalesOrderDetailsId); err != nil {
			tx.Rollback()
			return data, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
		}
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

	// Sales Order is the source for everything the new schedule/route/cost rows
	// below need that isn't already on the DR itself - customer name, ship-to
	// address, receiver/contact, ship type, doc no. The DR only carries
	// SalesOrderID, not these directly.
	var order models.Order
	if err := tx.First(&order, data.SalesOrderID).Error; err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed loading sales order for schedule/route")
	}

	// Changed 2026-09-03 (user decision): tbl_calendar_schedule used to get a
	// fresh row on every single DR ("Delivery Receipt #<DocNo>"), with no dedup
	// at all - a sales order delivered across 3 DRs got 3 identical-looking
	// calendar entries. Now it's created once per sales order ("Delivery to
	// <Customer>") - if one already exists for this SO, leave it alone rather
	// than creating another.
	var existingSchedule models.CalendarScheduleModel
	scheduleErr := tx.Where("reference_doc_id = ? AND department_type = ?", data.SalesOrderID, "Logistics").
		First(&existingSchedule).Error

	if errors.Is(scheduleErr, gorm.ErrRecordNotFound) {
		schedule := models.CalendarScheduleModel{
			ReferenceDocId: &data.SalesOrderID,
			CalendarScheduleContent: models.CalendarScheduleContent{
				DepartmentType: "Logistics",
				Title:          fmt.Sprintf("Delivery to %s", order.CustomerName),
				StartDate:      data.DeliveryDate,
				EndDate:        data.DeliveryDate,
				Description:    "",
			},
		}

		if _, _, err := s.CalendarScheduleService.CreateCalendarScheduleService(tx, &schedule, at); err != nil {
			tx.Rollback()
			return data, fiber.StatusInternalServerError, errors.New("failed creating calendar schedule")
		}
	} else if scheduleErr != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed checking for an existing calendar schedule")
	}

	// tbl_dispatching_logistics_calendar_schedule (department-specific, distinct
	// from the generic tbl_calendar_schedule above) is unchanged - still one per
	// DR, not deduped by sales order. Only the generic calendar entry above was
	// asked to stop duplicating.
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

	// Best-effort ship type name - ShipName stays "" if the id doesn't resolve,
	// same tolerance as getSalesOrderDocNo above.
	var shipType models.ShipType
	tx.First(&shipType, order.Ship_Type_ID)

	// Changed 2026-09-03 (user decision): every DR now also creates its own
	// route leg under that sales order's logistics schedule - always, even the
	// very first DR for that SO. Unlike the generic calendar schedule above,
	// this is never deduped: several deliveries against one SO are several
	// distinct routes (each with its own departed/arrived/returned times, filled
	// in later by the actual dispatch workflow - blank here at creation).
	// Receiver/ContactNo/CustomerName/ShipTo come from the sales order, not the
	// DR - the DR has no fields of its own for these.
	route := dispatching_models.LogisticsRoute{
		LogisticsRouteContent: dispatching_models.LogisticsRouteContent{
			ScheduleId:         logisticsSchedule.ID,
			ShipType:           shipType.ShipName,
			ReferenceDoc:       getSalesOrderDocNo(tx, data.SalesOrderID),
			DeliveryReceiptDoc: strconv.Itoa(data.DocNo),
			ClientSupplier:     order.CustomerName,
			Location:           order.ShipTo,
			Receiver:           order.Receiver,
			ContactNo:          order.ContactNo,
		},
	}

	if err := services.DbInsert(tx, &route); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating logistics route")
	}

	// This used to COPY the DR's cost breakdown onto the route as separate
	// LogisticsRouteCost rows - the mirror image of the copy that ran when a route
	// was saved, which is how the two tables drifted apart.
	//
	// User decision, 2026-09-03: there is one delivery-cost table. The DR's cost rows
	// were inserted above against this receipt; attaching the new route is now just
	// stamping route_id on them, so the route form and the DR read the identical
	// rows rather than two copies (§13.3, "entered once").
	if err := tx.Model(&dispatching_models.DeliveryReceiptCosts{}).
		Where("delivery_receipt_id = ? AND route_id = 0", data.ID).
		Update("route_id", route.ID).Error; err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed attaching delivery costs to logistics route")
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
		if err := NormalizeDeliveryReceiptCosts(tx, update.DeliveryReceiptCosts); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed preparing delivery costs")
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
	// Changed 2026-09-03, same as create: title is "Delivery to <Customer>", not
	// "Delivery Receipt #<DocNo>" - was still overwriting back to the old format
	// on every edit, which would have silently undone the create-time rename the
	// moment anyone touched a DR against that sales order.
	genericResult := tx.Model(&models.CalendarScheduleModel{}).
		Where("reference_doc_id = ? AND department_type = ?", receipt.SalesOrderID, "Logistics").
		Updates(map[string]interface{}{
			"start_date": update.DeliveryDate,
			"end_date":   update.DeliveryDate,
			"title":      fmt.Sprintf("Delivery to %s", receipt.CustomerName),
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
				Title:          fmt.Sprintf("Delivery to %s", receipt.CustomerName),
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

	receipt, status, err := s.GetDeliveryReceiptService(conditions)
	if err != nil {
		tx.Rollback()
		return receipt, status, errors.New("delivery receipt not found")
	}

	// User decision, 2026-09-03: deleting a DR also removes the logistics schedule
	// and route it created, and every cost row keyed to it. Nothing here relies on
	// an ON DELETE CASCADE - those FKs are AutoMigrate-created and are not guaranteed
	// to exist on a restored database, which is why the children are removed by hand
	// and in dependency order.
	//
	// Scope is what this DR brought into being: schedules carrying its id, and their
	// routes. A route in somebody else's schedule that merely types this DR's number
	// into its DELIVERY RECEIPT field is left standing - it is a trip the dispatcher
	// built, not one the receipt created - though its cost rows go with the receipt
	// below, since those were the receipt's delivery costs.
	var scheduleIDs []uint
	if err := tx.Model(&dispatching_models.LogisticsCalendarScheduleModel{}).
		Where("delivery_receipt_id = ?", receipt.ID).
		Pluck("id", &scheduleIDs).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed finding logistics schedules for this receipt")
	}

	for _, scheduleID := range scheduleIDs {
		if err := DeleteRouteCostsForSchedule(tx, scheduleID); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed deleting route costs")
		}
	}
	if len(scheduleIDs) > 0 {
		if err := tx.Where("schedule_id IN ?", scheduleIDs).
			Delete(&dispatching_models.LogisticsRoute{}).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed deleting logistics routes")
		}
		if err := tx.Where("id IN ?", scheduleIDs).
			Delete(&dispatching_models.LogisticsCalendarScheduleModel{}).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed deleting logistics schedules")
		}
	}

	// The generic tbl_calendar_schedule entry is deliberately NOT touched: it is one
	// per sales order, shared by every DR against that order, so removing it here
	// would erase the calendar entry for deliveries that still exist.

	// Receipt files hang off the cost rows, so they go first.
	var costIDs []uint
	if err := tx.Model(&dispatching_models.DeliveryReceiptCosts{}).
		Where("delivery_receipt_id = ?", receipt.ID).
		Pluck("id", &costIDs).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed finding delivery costs")
	}
	if len(costIDs) > 0 {
		if err := tx.Where("delivery_receipt_cost_id IN ?", costIDs).
			Delete(&dispatching_models.ReceiptFile{}).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed deleting receipt files")
		}
	}
	if err := tx.Where("delivery_receipt_id = ?", receipt.ID).
		Delete(&dispatching_models.DeliveryReceiptCosts{}).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed deleting delivery costs")
	}

	// Capture the SO lines before the items go, so their status can be recomputed
	// afterwards - a deleted DR means those units are no longer delivered (§7.1).
	var soDetailIDs []uint
	if err := tx.Model(&dispatching_models.DeliveryReceiptItems{}).
		Where("delivery_receipt_id = ? AND sales_order_details_id IS NOT NULL", receipt.ID).
		Distinct().Pluck("sales_order_details_id", &soDetailIDs).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed reading delivery receipt items")
	}
	if err := tx.Where("delivery_receipt_id = ?", receipt.ID).
		Delete(&dispatching_models.DeliveryReceiptItems{}).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed deleting delivery receipt items")
	}

	if err := services.DbDelete(tx, &receipt, conditions); err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed deleting delivery receipt")
	}

	for _, id := range soDetailIDs {
		if err := services.RecomputeSoItemStatus(tx, id); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
		}
	}

	atdata := dispatching_models.DeliveryReceiptAt{RefId: receipt.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed creating receipt audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	InvalidateDRCaches()

	return receipt, fiber.StatusOK, nil
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
