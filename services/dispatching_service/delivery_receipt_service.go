package dispatching_services

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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

// The customer's name for a calendar title or a route's CLIENT/SUPPLIER.
//
// tbl_trans_sales_order carries a customer_name column, but it is a denormalized
// copy that is not kept in step with the BPI record and is empty on every order
// in production - which is how logistics schedules came to be titled "Delivery
// to " with nothing after it (user-reported 2026-09-08). The Delivery Receipt
// stores its own copy and that one IS populated, so it is preferred; BPI is the
// authority behind both and settles it when neither copy has anything, resolving
// customer_id -> bpi.name exactly as the schedule details screen does.
func resolveCustomerName(tx *gorm.DB, stored string, customerID uint) string {
	if name := strings.TrimSpace(stored); name != "" {
		return name
	}

	if customerID != 0 {
		var bpi models.Bpi
		if err := tx.First(&bpi, customerID).Error; err == nil {
			return strings.TrimSpace(bpi.Name)
		}
	}

	return ""
}

// Dates in the calendar and delivery-receipt tables are free-text nvarchar, and
// two writers store two different shapes: the calendar UI writes ISO
// ("2026-09-04T00:00:00") while the Delivery Receipt copies its own
// delivery_date straight through ("09/04/2026"). Grouping schedules by day
// therefore has to compare the day itself and not the string - matching on the
// raw text would miss every schedule written by the other writer and hand each
// DR its own schedule again, which is the behaviour the grouping exists to stop.
func parseScheduleDay(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	// "01/02/2006" is Go's reference layout for MM/dd/yyyy - what an en-PH client
	// sends, that culture's short-date pattern being M/d/yyyy.
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
		"1/2/2006",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), true
		}
	}

	return time.Time{}, false
}

// §13.3 gives a logistics schedule 1..n routes, and the 2026-09-03 decision put
// every DR's route "under that sales order's logistics schedule". The code did
// not do that: it built a fresh schedule for each DR and attached the route to
// the new one, so two DRs against one sales order became two one-route schedules
// instead of one two-route schedule - and opening either calendar entry showed a
// single route (user-reported 2026-09-08).
//
// The grouping key is sales order AND delivery day, per the user's decision:
// two deliveries on one order on different days are separate trips and keep
// separate schedules. Only the day is compared - see parseScheduleDay.
//
// delivery_receipt_id and delivery_receipt_doc_no are deliberately left unset
// here (user decision, 2026-09-08). They are single-valued columns on a row that
// now covers several receipts, so whatever they held would name one arbitrary DR
// out of the set; each route carries its own delivery_receipt_doc instead.
func (s *DeliveryReceiptService) findOrCreateLogisticsSchedule(
	tx *gorm.DB,
	salesOrderID uint,
	deliveryDate string,
	customerName string,
	at models.At,
) (*dispatching_models.LogisticsCalendarScheduleModel, error) {
	title := fmt.Sprintf("Delivery to %s", customerName)

	if wanted, ok := parseScheduleDay(deliveryDate); ok {
		var existing []dispatching_models.LogisticsCalendarScheduleModel
		if err := tx.Where("sales_order_id = ? AND department = ?", salesOrderID, "LOGISTICS").
			Find(&existing).Error; err != nil {
			return nil, err
		}

		// Sorted here rather than with Order("id") on the query. tx is the
		// caller's transaction handle, and an ORDER BY chained onto it survives
		// into the next statement built from the same handle - the very next one
		// is CreateLogisticsSchedule's driver lookup, which adds its own
		// ordering and then fails on "a column has been specified more than once
		// in the order by list". Ordering the slice keeps the oldest schedule
		// winning - which matters where rows already exist for the same order
		// and day, exactly the split this repairs - without touching the shared
		// statement.
		sort.Slice(existing, func(a, b int) bool {
			return existing[a].ID < existing[b].ID
		})

		for i := range existing {
			day, ok := parseScheduleDay(existing[i].StartDate)
			if !ok || !day.Equal(wanted) {
				continue
			}

			// Correct a title left behind by the old per-DR naming
			// ("Delivery Receipt #2"): the schedule now covers every receipt
			// delivered to this customer on this day, so a single receipt
			// number in the title names only one of them.
			if existing[i].Title != title {
				if err := tx.Model(&dispatching_models.LogisticsCalendarScheduleModel{}).
					Where("id = ?", existing[i].ID).
					Update("title", title).Error; err != nil {
					return nil, err
				}
				existing[i].Title = title
			}

			return &existing[i], nil
		}
	}

	schedule := dispatching_models.LogisticsCalendarScheduleModel{
		CalendarScheduleBase: dispatching_models.CalendarScheduleBase{
			Department: "LOGISTICS",
			StartDate:  deliveryDate,
			EndDate:    deliveryDate,
			Title:      title,
		},
		LogisticsCalendarScheduleContent: dispatching_models.LogisticsCalendarScheduleContent{
			SalesOrderId:    salesOrderID,
			SalesOrderDocNo: getSalesOrderDocNo(tx, salesOrderID),
		},
	}

	if _, _, err := s.LogisticsCalendarScheduleService.CreateLogisticsSchedule(tx, &schedule, at); err != nil {
		return nil, err
	}

	return &schedule, nil
}

// Removes a logistics schedule that this service created and that now holds no
// routes - the case being a delivery date change moving its last route to
// another day, which would otherwise leave an empty card sitting on the old
// date. Anything a dispatcher entered by hand (people, vehicle, notes,
// description) means the schedule is theirs rather than ours, so it stays.
//
// Mirrors DeleteLogisticsSchedule's delete-plus-audit, inlined because that one
// opens its own transaction and this runs inside the DR's.
func deleteEmptyLogisticsSchedule(tx *gorm.DB, scheduleID uint, at models.At) error {
	if scheduleID == 0 {
		return nil
	}

	var routeCount int64
	if err := tx.Model(&dispatching_models.LogisticsRoute{}).
		Where("schedule_id = ?", scheduleID).
		Count(&routeCount).Error; err != nil {
		return err
	}
	if routeCount > 0 {
		return nil
	}

	var schedule dispatching_models.LogisticsCalendarScheduleModel
	if err := tx.First(&schedule, scheduleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if schedule.VehicleId != 0 ||
		strings.TrimSpace(schedule.DriverName) != "" ||
		strings.TrimSpace(schedule.People) != "" ||
		strings.TrimSpace(schedule.Notes) != "" ||
		strings.TrimSpace(schedule.Description) != "" {
		return nil
	}

	if err := tx.Delete(&dispatching_models.LogisticsCalendarScheduleModel{}, scheduleID).Error; err != nil {
		return err
	}

	audit := dispatching_models.LogisticsCalendarScheduleModelAt{
		CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
			RefId: scheduleID,
			At:    at,
		},
	}
	return services.DbInsert(tx, &audit)
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
				Title:          fmt.Sprintf("Delivery to %s", resolveCustomerName(tx, data.CustomerName, data.CustomerID)),
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
	// from the generic tbl_calendar_schedule above) now groups by sales order and
	// delivery day rather than getting a fresh row per DR - several receipts
	// delivered to one customer on one day share one schedule and appear on it as
	// separate routes. See findOrCreateLogisticsSchedule.
	logisticsSchedule, scheduleErr2 := s.findOrCreateLogisticsSchedule(
		tx, data.SalesOrderID, data.DeliveryDate, resolveCustomerName(tx, data.CustomerName, data.CustomerID), at)
	if scheduleErr2 != nil {
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
			ClientSupplier:     resolveCustomerName(tx, data.CustomerName, data.CustomerID),
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
			"title":      fmt.Sprintf("Delivery to %s", resolveCustomerName(tx, receipt.CustomerName, receipt.CustomerID)),
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
				Title:          fmt.Sprintf("Delivery to %s", resolveCustomerName(tx, receipt.CustomerName, receipt.CustomerID)),
				StartDate:      update.DeliveryDate,
				EndDate:        update.DeliveryDate,
			},
		}
		if _, _, err := s.CalendarScheduleService.CreateCalendarScheduleService(tx, &schedule, at); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed creating calendar schedule")
		}
	}

	// The logistics schedule is no longer keyed to a single DR - delivery_receipt_id
	// is deliberately left unset now that one schedule covers every receipt
	// delivered to a customer on one day - so this can no longer find its schedule
	// by that column. It resolves the schedule the DR belongs to *now*, from its
	// sales order and its (possibly edited) delivery date, and moves this DR's
	// route onto it.
	//
	// Re-parenting the route rather than editing the old schedule's dates matters:
	// that schedule may carry other receipts' routes, and dragging its date along
	// to follow this one receipt would move all of them.
	targetSchedule, targetErr := s.findOrCreateLogisticsSchedule(
		tx, receipt.SalesOrderID, update.DeliveryDate, resolveCustomerName(tx, receipt.CustomerName, receipt.CustomerID), at)
	if targetErr != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed syncing logistics calendar schedule")
	}

	var orderSchedules []dispatching_models.LogisticsCalendarScheduleModel
	if err := tx.Where("sales_order_id = ? AND department = ?", receipt.SalesOrderID, "LOGISTICS").
		Find(&orderSchedules).Error; err != nil {
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed loading logistics calendar schedules")
	}

	// The route is this DR's leg. delivery_receipt_doc is the link (routes carry
	// no receipt id), scoped to the schedules of this sales order so a doc number
	// reused under a different order cannot match.
	//
	// Two statements rather than one with a subquery: a subquery built from tx
	// chains onto the same statement the outer query uses, and this handle has
	// already been shown to carry clauses across executions (see the sort in
	// findOrCreateLogisticsSchedule). Fetching the ids first keeps each statement
	// to the plain Model/Where/execute shape used everywhere else here.
	var scheduleIDs []uint
	for _, sched := range orderSchedules {
		scheduleIDs = append(scheduleIDs, sched.ID)
	}

	var route dispatching_models.LogisticsRoute
	routeErr := gorm.ErrRecordNotFound
	if len(scheduleIDs) > 0 {
		routeErr = tx.Where(
			"delivery_receipt_doc = ? AND schedule_id IN ?",
			strconv.Itoa(receipt.DocNo),
			scheduleIDs,
		).First(&route).Error
	}

	switch {
	case routeErr == nil:
		previousScheduleID := route.ScheduleId
		if err := tx.Model(&dispatching_models.LogisticsRoute{}).
			Where("id = ?", route.ID).
			Updates(map[string]interface{}{
				"schedule_id":   targetSchedule.ID,
				"reference_doc": getSalesOrderDocNo(tx, receipt.SalesOrderID),
			}).Error; err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed moving logistics route")
		}

		if previousScheduleID != targetSchedule.ID {
			if err := deleteEmptyLogisticsSchedule(tx, previousScheduleID, at); err != nil {
				tx.Rollback()
				return receipt, fiber.StatusInternalServerError, errors.New("failed tidying empty logistics schedule")
			}
		}

	case errors.Is(routeErr, gorm.ErrRecordNotFound):
		// This DR predates the calendar-linking feature and never got a route.
		// Give it one now rather than leaving a schedule with nothing on it.
		var order models.Order
		tx.First(&order, receipt.SalesOrderID)

		var shipType models.ShipType
		tx.First(&shipType, order.Ship_Type_ID)

		newRoute := dispatching_models.LogisticsRoute{
			LogisticsRouteContent: dispatching_models.LogisticsRouteContent{
				ScheduleId:         targetSchedule.ID,
				ShipType:           shipType.ShipName,
				ReferenceDoc:       getSalesOrderDocNo(tx, receipt.SalesOrderID),
				DeliveryReceiptDoc: strconv.Itoa(receipt.DocNo),
				ClientSupplier:     resolveCustomerName(tx, receipt.CustomerName, receipt.CustomerID),
				Location:           order.ShipTo,
				Receiver:           order.Receiver,
				ContactNo:          order.ContactNo,
			},
		}
		if err := services.DbInsert(tx, &newRoute); err != nil {
			tx.Rollback()
			return receipt, fiber.StatusInternalServerError, errors.New("failed creating logistics route")
		}

	default:
		tx.Rollback()
		return receipt, fiber.StatusInternalServerError, errors.New("failed locating logistics route")
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
