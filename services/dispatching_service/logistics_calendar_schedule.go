package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type LogisticsCalendarScheduleService struct{}

func NewLogisticsCalendarScheduleService() *LogisticsCalendarScheduleService {
	return &LogisticsCalendarScheduleService{}
}

func (s *LogisticsCalendarScheduleService) GetLogisticsSchedules(conditions map[string]interface{}) (*[]dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	var schedules = &[]dispatching_models.LogisticsCalendarScheduleModel{}

	if err := initializers.DB.Where(conditions).
		Preload("Routes").Preload("Routes.Costs").
		Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, fiber.StatusInternalServerError, errors.New("failed getting logistics calendar schedules")
	}

	return schedules, fiber.StatusOK, nil
}

// GET a single logistics schedule by ID
func (s *LogisticsCalendarScheduleService) GetLogisticsSchedule(conditions map[string]interface{}) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	var schedule = &dispatching_models.LogisticsCalendarScheduleModel{}

	if err := initializers.DB.Where(conditions).
		Preload("Routes").Preload("Routes.Costs").
		First(schedule).Error; err != nil {
		return schedule, fiber.StatusNotFound, errors.New("logistics calendar schedule not found")
	}

	return schedule, fiber.StatusOK, nil
}

// CREATE a logistics schedule within the caller's transaction, so it commits/rolls
// back automically with whatever operation is creating it (e.g. a Delivery Receipt).
// It does not begin or commit a transaction itself.
func (s *LogisticsCalendarScheduleService) CreateLogisticsSchedule(tx *gorm.DB, schedule *dispatching_models.LogisticsCalendarScheduleModel, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	if err := services.DbInsert(tx, &schedule); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return schedule, fiber.StatusInternalServerError, errors.New("duplicate record error")
		}
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule")
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	return schedule, fiber.StatusOK, nil
}

// UpdateLogisticsSchedule updates within the caller's transaction so it commits/rolls
// back atomically with whatever operation is updating it (e.g. a Delivery Receipt).
// It does not begin or commit a transaction itself.
func (s *LogisticsCalendarScheduleService) UpdateLogisticsSchedule(tx *gorm.DB, schedule *dispatching_models.LogisticsCalendarScheduleModel, conditions map[string]interface{}, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed updating logistics schedule")
	}

	// Replace routes (Internal only) with the incoming set. DbUpdate only touches
	// the schedule's own columns, not nested associations, so this is manual —
	// delete-then-reinsert, same pattern as DeliveryReceipt's items/costs.
	if err := tx.Where("schedule_id = ?", schedule.ID).Delete(&dispatching_models.LogisticsRoute{}).Error; err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed clearing old routes")
	}
	if len(schedule.Routes) > 0 {
		for i := range schedule.Routes {
			schedule.Routes[i].ID = 0
			schedule.Routes[i].ScheduleId = schedule.ID
			for j := range schedule.Routes[i].Costs {
				schedule.Routes[i].Costs[j].ID = 0
			}
		}
		if err := tx.Create(&schedule.Routes).Error; err != nil {
			return schedule, fiber.StatusInternalServerError, errors.New("failed saving routes")
		}

		// §7.1 rows 16-17 (FAILED/RETURN, DELIVERED) - a route's departed_at/arrived_at/
		// returned_at is what actually distinguishes those two states, so every SO line
		// on whichever Delivery Receipt this route is tracking needs recomputing here.
		for _, route := range schedule.Routes {
			if err := services.RecomputeSoItemStatusForDeliveryReceiptDoc(tx, route.DeliveryReceiptDoc); err != nil {
				return schedule, fiber.StatusInternalServerError, errors.New("failed recomputing SO item status")
			}
		}
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	return schedule, fiber.StatusOK, nil
}

// DELETE a logistics schedule
func (s *LogisticsCalendarScheduleService) DeleteLogisticsSchedule(conditions map[string]interface{}, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.LogisticsCalendarScheduleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	schedule, status, err := s.GetLogisticsSchedule(conditions)
	if err != nil {
		return schedule, status, errors.New("logistics calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed deleting logistics calendar schedule")
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}
